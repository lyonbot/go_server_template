import fs from 'fs';
import { defineConfig } from '@rsbuild/core';
import { pluginVue } from '@rsbuild/plugin-vue';
import { pluginSass } from '@rsbuild/plugin-sass';
import UnoCSS from '@unocss/postcss'
import { web as webDeploymentConfig } from '../.deployment.cjs'

const IS_PROD = process.env.NODE_ENV === 'production';
const CDN_PUBLIC_PATH = webDeploymentConfig.cdn?.publicPath;

const entries = fs.readdirSync('./src/entries')

export default defineConfig({
  source: {
    entry: Object.fromEntries(
      entries.map(entry => [entry, [`./src/env.ts`, `./src/entries/${entry}/index.ts`]])
    ),
  },
  plugins: [
    pluginSass(),
    pluginVue(),
  ],
  output: {
    assetPrefix: (IS_PROD && CDN_PUBLIC_PATH) || './',
    cleanDistPath: IS_PROD,
    distPath: { root: './dist' }
  },
  server: {
    base: '/web',
    compress: false,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: false,
      },
    }
  },
  html: {
    template: './index.html',
  },
  tools: {
    postcss: {
      postcssOptions: {
        plugins: [
          UnoCSS(),
          require('postcss-import'), // inline @import
          require('postcss-preset-env')({
            stage: 2,
            features: {
              'oklab-function': true,
            },
          }),
        ],
      }
    },
  },
});
