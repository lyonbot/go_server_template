## go server template

### 初始化

仓库内全局替换 my_app 为你的项目名称。

然后

```sh
cp deployment.cjs.example deployment.cjs
cp .env.example .env

# 修改 deployment.cjs
# 修改 .env

cp .env .env.production   # 注意：多 .env 不会合并
pnpm install
```

远端的机器可能需要这些准备

- 安装 pm2
- go env -w GOPROXY=https://goproxy.cn,direct

首次部署完成，需要到服务器 `pm2 save` 一下，避免服务器重启后 pm2 丢失配置。

### 构建

为了避免环境差异，go 构建工作会放到目标机器运行；web 构建则会在本地运行。

本地环境需要

- nodejs>=20 & pnpm@9 用于运行 ./web 下的 npm run build

远端机器如果有墙可以考虑：

```sh
go env -w GOPROXY=https://goproxy.cn,direct
```

### 环境变量

环境变量需要在 ./dist/ecosystem.config.js 设置。本地开发则会加载 ./.env

- PORT: 服务端口，例如：`8080`
- JWT_SECRET: 用于生成 JWT 的密钥，例如：`50450000-dead-beef-1234-7ee4f3e70000`
- REDIS_ADDR: Redis 地址，例如：`127.0.0.1:6379`

### 一些工具

- 生成色板 https://uicolors.app/create 然后写到 ./web/uno.config.ts 中
