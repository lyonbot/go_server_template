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

  # 检查需要 copy 的环境变量文件
  # (根据 ecosystem.config.js 中的 @import: 注释来判断) 
  env_files_to_copy=$(grep -Po '(?<=@import:)\s*\S+' $DEPLOY_FILE_DIR/ecosystem.config.js | sort | uniq)
  if [ -n "$env_files_to_copy" ]; then
    # 确保每一个环境变量都在 production 存在
    base_env_file_fields="$(awk 'match($0, /^[_A-Za-z0-9]+/){ print substr($0, RSTART, RLENGTH) }' .env)"
    for env_file_name in $env_files_to_copy; do
      for env_field_name in $base_env_file_fields; do
        if ! grep -q "^$env_field_name=" $env_file_name; then
          echo "Error: env $env_field_name is not declared in $env_file_name"
          exit 1
        fi
      done
    done
    # 检查通过即可 rsync
    rsync -avL ${env_files_to_copy} ${DEST_HOST}:${APP_DIR}/
  fi

  # 构建web
  build_web_if_needed() {
    if $NO_WEB; then
      echo "skipped: build web"
    elif [ "$(ls -1t web | head -n 1)" = "dist" ]; then
      echo "web/dist is already newest"
    else
      npm run build:web
    fi
  }
  (build_web_if_needed && rsync -avL --delete ./web/dist ${DEST_HOST}:${SOURCE_DIR}/web/) &

  git push $IS_FORCE ssh://${DEST_HOST}/${SOURCE_DIR} HEAD:staging || {
    wait
    echo "你可能需要先去服务器 mkdir -p $SOURCE_DIR && cd $SOURCE_DIR && git init" 1>&2
    exit 1
  }

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
cp -L -r $DEPLOY_FILE_DIR/* $APP_DIR/

cd $APP_DIR
node $SOURCE_DIR/deploy/hooks/before-update-pm2.js
pm2 reload ./ecosystem.config.js

echo "部署完成"
