// Hydrate the i18n locale on the client side:
// 1. If authenticated and profile has a language → use it
// 2. Else if localStorage has a saved locale → use it
// 3. Else fall back to the module default ('en')

const LOCALE_KEY = 'colloquia.locale'

export default defineNuxtPlugin(async () => {
  const nuxtApp = useNuxtApp()
  const { isAuthenticated, getProfile } = useAuth()

  if (isAuthenticated.value) {
    try {
      const profile = await getProfile()
      if (profile.language) {
        await nuxtApp.$i18n.setLocale(profile.language)
        localStorage.setItem(LOCALE_KEY, profile.language)
        return
      }
    } catch {
      // profile fetch failed — fall through to localStorage
    }
  }

  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved) {
    try { await nuxtApp.$i18n.setLocale(saved as 'en' | 'pt') } catch { /* unknown code — ignore */ }
  }
})
