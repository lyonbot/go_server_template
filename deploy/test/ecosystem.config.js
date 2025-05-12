module.exports = {
  apps: [{
    name: 'SERVICE_NAME',  // will be replaced by deploy.sh
    script: './main',
    env_file: '.env.production',
    env: {
      NODE_ENV: 'production',
      GIN_MODE: 'release',
      /* @import: .env.production */
    },
    watch: false,
    // watch: ['./main'],
  }],
};