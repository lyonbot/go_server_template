#!/bin/bash

# 这个文件包含2部分：如果在本地执行，会自动部署到服务器；如果在服务器执行，会自动构建并重启服务

set -e
IS_FORCE=
if echo "$@" | grep -q -- "-f"; then IS_FORCE="-f"; fi

# ---------------------------
# following code runs on local machine
# (if args not containing "--is-remote" )

if [ "$1" != "--is-remote" ]; then
  $(node -e 'require("./.deployment.cjs").printGoEnv()')

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
  (npm run build:web && rsync -avL --delete ./web/dist ${DEST_HOST}:${SOURCE_DIR}/web/) &

  git push $IS_FORCE ssh://${DEST_HOST}${SOURCE_DIR} HEAD:staging || {
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
  
  ssh $DEST_HOST "export SOURCE_DIR=\"$SOURCE_DIR\" APP_DIR=\"$APP_DIR\" SERVICE_NAME=\"$SERVICE_NAME\"; cd $SOURCE_DIR && $MERGE_CMD && bash -le ./deploy.sh --is-remote $*"
  exit $?
fi

# ---------------------------
# following code runs on target machine

echo "开始构建... $SERVICE_NAME"
git merge staging
git branch -D staging
go build -ldflags "-s -w" -tags release -o $APP_DIR/main ./cmd/main
cp -r dist/* $APP_DIR/
sed -i -e "s/SERVICE_NAME/$SERVICE_NAME/g" $APP_DIR/ecosystem.config.js

# 确保构建成功
if [ $? -ne 0 ]; then
    echo "构建失败"
    exit 1
fi

cd $APP_DIR && pm2 reload ./ecosystem.config.js
echo "部署完成"
