import { defineConfig, presetIcons, presetWind3, transformerDirectives } from 'unocss'

// 生成色板 https://uicolors.app/create

export default defineConfig({
  content: {
    filesystem: [
      '**/*.{html,js,ts,jsx,tsx,vue,svelte,astro}',
    ],
  },
  presets: [
    presetWind3({
      dark: 'media',
    }),
    presetIcons({
      extraProperties: {
        display: 'inline-block',
        'vertical-align': 'middle',
        // ...
      },
      collections: {
        mdi: () => import('@iconify-json/mdi/icons.json').then(i => i.default),
      },
    }),
  ],
  transformers: [
    transformerDirectives(),
  ],
  // 一些快捷类名
  shortcuts: [
    [/^vflex(-\d+|-\[.+)?$/, ([, c]) => `flex flex-col gap${c || '-4'}`],
    ['flex-center', 'flex justify-center items-center'],
  ],
  theme: {
    colors: {
      brand: {
        DEFAULT: '#8c42d7',
        '50': '#faf6fe',
        '100': '#f2eafd',
        '200': '#e8d9fb',
        '300': '#d6bbf7',
        '400': '#bd90f0',
        '500': '#a465e7',
        '600': '#8c42d7',
        '700': '#7933bd',
        '800': '#672f9a',
        '900': '#55277c',
        '950': '#38105b',
      },
    }
  }
})