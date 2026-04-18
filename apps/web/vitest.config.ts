import { defineVitestConfig } from '@nuxt/test-utils/config'

export default defineVitestConfig({
  test: {
    environment: 'nuxt',
    environmentOptions: {
      nuxt: {
        mock: {
          intersectionObserver: true,
          indexedDb: true,
        },
      },
    },
    globals: true,
    coverage: {
      provider: 'v8',
      include: ['app/composables/**', 'app/middleware/**'],
      exclude: ['app/composables/__tests__/**', 'app/middleware/__tests__/**'],
    },
  },
})
