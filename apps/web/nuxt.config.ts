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
  modules: ['@nuxtjs/tailwindcss'],
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