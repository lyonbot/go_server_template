#!/bin/bash

# 这个文件包含2部分：如果在本地执行，会自动部署到服务器；如果在服务器执行，会自动构建并重启服务

set -e

ALL_ARGS="$*"

DEPLOY_TYPE=
IS_FORCE=
IS_ON_REMOTE=false
DESTINATION_DIR=
NO_WEB=false      # 是否不构建 web

while [ $# -gt 0 ]; do
  case "$1" in
    -w|--no-web)
      NO_WEB=true
      ;;
    -f|--force)
      IS_FORCE="-f"
      ;;
    --is-remote)
      IS_ON_REMOTE=true
      ;;
    *)
      if [ -z "$DEPLOY_TYPE" ]; then
        DEPLOY_TYPE=$1
      else
        echo "Error: too many arguments"
        exit 1
      fi
      ;;
  esac
  shift
done

DEPLOY_TYPE=${DEPLOY_TYPE:-test}  # default to test

# ---------------------------
# following code runs on local machine
# (if args not containing "--is-remote" )

if ! $IS_ON_REMOTE; then
  echo "DEPLOY_TYPE: $DEPLOY_TYPE"
  deploy_env="$(node -e "require('./.deployment.cjs').printGoEnv('$DEPLOY_TYPE')")"
  export $deploy_env
  echo 开始推送... $deploy_env

  # 必须提交了 git
  if [ -n "$(git status --porcelain)" ]; then
      echo "Error: You have uncommitted changes. Please commit or stash them first."
      exit 1
  fi

  # 确保每一个环境变量都在 production 存在
  awk 'match($0, /^[_A-Za-z0-9]+/){ print substr($0, RSTART, RLENGTH) }' .env | while read -r env; do
    if ! grep -q "^$env=" .env.production; then
      echo "Error: env $env is not declared in .env.production"
      exit 1
    fi
  done

  # 开始构建
  (( $NO_WEB || npm run build:web ) && rsync -avL --delete ./web/dist ${DEST_HOST}:${SOURCE_DIR}/web/) &

  git push $IS_FORCE ssh://${DEST_HOST}/${SOURCE_DIR} HEAD:staging || {
    wait
    echo "你可能需要先去服务器 mkdir -p $SOURCE_DIR && cd $SOURCE_DIR && git init" 1>&2
    exit 1
  }
  rsync -avL ./.env.production ${DEST_HOST}:${APP_DIR}/

  wait

  MERGE_CMD="git merge staging"
  if [ -n "$IS_FORCE" ]; then
    echo "强制更换分支内容"
    MERGE_CMD="git reset --hard staging"
  fi

  ssh $DEST_HOST "export $deploy_env; cd $SOURCE_DIR && $MERGE_CMD && bash -le ./deploy.sh --is-remote $ALL_ARGS"
  exit $?
fi

# ---------------------------
# following code runs on target machine

echo "开始构建... $SERVICE_NAME"
git merge staging
git branch -D staging
go build -ldflags "-s -w" -tags release -o $APP_DIR/main ./cmd/main
cp -r $DEPLOY_FILE_DIR/* $APP_DIR/

node -e '
const fs = require("fs");
const { SERVICE_NAME, APP_DIR } = process.env;

var pm2configPath = `${APP_DIR}/ecosystem.config.js`;
var content = fs.readFileSync(pm2configPath, "utf8");

content = content.replace(/SERVICE_NAME/g, SERVICE_NAME);
content = content.replace(/\/\*+ @import:\s*(\S+)\s*\*\//g, (_, envFilePath) => {
  var envFileContent = fs.readFileSync(envFilePath, "utf8");
  var output = [_];
  envFileContent.split("\n").forEach(line => {
    line = line.trim();
    if (!line.startsWith("#") && line.contains("=")) {
      var idx = line.indexOf("=");
      var key = line.substring(0, idx).trim();
      var value = line.substring(idx + 1).trim();
      output.push(`      ${JSON.stringify(key)}: ${JSON.stringify(value)},`);
    }
  });
  return output.join("\n");
});

fs.writeFileSync(pm2configPath, content);
'

# 确保构建成功
if [ $? -ne 0 ]; then
    echo "构建失败"
    exit 1
fi

cd $APP_DIR && pm2 reload ./ecosystem.config.js
echo "部署完成"
