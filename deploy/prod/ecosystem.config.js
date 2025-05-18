module.exports = {
  apps: [{
    name: 'SERVICE_NAME',  // will be replaced by hooks/before-update-pm2.js
    script: './main',
    env: {
      NODE_ENV: 'production',
      GIN_MODE: 'release',

      // 导入环境变量文件
      /* @import: .env.production */
    },
    watch: false,
    // watch: ['./main'],
  }],
};