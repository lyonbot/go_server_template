## go server template

### 初始化

仓库内全局替换 my_app 为你的项目名称。

然后

```sh
cp deployment.cjs.example deployment.cjs
cp .env.example .env

# 修改 deployment.cjs
# 修改 .env

cp .env .env.test   # 注意：多 .env 不会合并
pnpm install
```

远端的机器可能需要这些准备（以 Ubuntu 为例）：

```sh
# 配置 nodejs 和 golang 源
curl -sL https://deb.nodesource.com/setup_20.x -o /tmp/nodesource_setup.sh
bash /tmp/nodesource_setup.sh
sudo add-apt-repository ppa:longsleep/golang-backports

# 安装 nodejs 和 golang
sudo apt install nodejs golang-go

# 配置镜像
go env -w GOPROXY=https://goproxy.cn,direct
npm config set registry https://registry.npmmirror.com

# 安装 pm2 和 pnpm
sudo npm install -g pnpm@10 pm2
pm2 startup
```

首次部署完成，需要到服务器 `pm2 save` 一下，避免服务器重启后 pm2 丢失配置。

### 构建

为了避免环境差异，go 构建工作会放到目标机器运行；web 构建则会在本地运行。

本地环境需要

- nodejs>=20 & pnpm@10 用于运行 ./web 下的 npm run build

### 环境变量

本地开发则会加载 `./.env`

- PORT: 服务端口，例如：`8080`
- JWT_SECRET: 用于生成 JWT 的密钥，例如：`50450000-dead-beef-1234-7ee4f3e70000`
- REDIS_ADDR: Redis 地址，例如：`127.0.0.1:6379`

远端部署的则参考 `./.deployment.cjs` 以及 `./deploy/*/ecosystem.config.js`

### 一些工具

- 生成色板 https://uicolors.app/create 然后写到 ./web/uno.config.ts 中
