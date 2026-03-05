import { defineNuxtConfig } from 'nuxt/config'

export default defineNuxtConfig({
  modules: [
    '@pinia/nuxt',
    '@nuxtjs/tailwindcss',
  ],

  // Use this instead of a separate postcss.config.cjs
  postcss: {
    plugins: {
      // Add custom PostCSS plugins here
      'postcss-nested': {},
      autoprefixer: {}, // optional, Tailwind already includes this
    },
  },
})