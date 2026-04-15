// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  ssr: false,
  devtools: { enabled: true },
  vue: {
    compilerOptions: {
      isCustomElement: (tag) => tag === 'emoji-picker',
    },
  },
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/i18n'],
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'en',
    locales: [
      { code: 'en', name: 'English', file: 'en.json' },
      { code: 'pt', name: 'Português', file: 'pt.json' },
    ],
    langDir: 'locales/',
    detectBrowserLanguage: false,
  },
  tailwindcss: {
    cssPath: '~/assets/css/main.css',
  },
  runtimeConfig: {
    apiBase: 'http://localhost:8000',
    public: {
      apiBase: 'http://localhost:8000',
    },
  },
})