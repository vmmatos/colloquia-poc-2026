const LOCALE_KEY = 'colloquia.locale'

export function useLocale() {
  const { locale, locales } = useI18n()
  const nuxtApp = useNuxtApp()
  const { auth } = useAuth()

  const availableLocales = computed(() =>
    (locales.value as Array<{ code: string; name: string }>).map(l => ({
      code: l.code,
      name: l.name,
    }))
  )

  async function setLocale(code: string) {
    // Apply immediately — reactive update for all $t() in the app
    await nuxtApp.$i18n.setLocale(code as 'en' | 'pt')

    // Persist to localStorage as fallback for pre-login / offline
    if (import.meta.client) {
      localStorage.setItem(LOCALE_KEY, code)
    }

    // Persist to user profile (best-effort — don't block UI on failure)
    if (auth.value.access_token) {
      $fetch('/api/users/me', {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
        body: { language: code },
      }).catch(() => { /* silent — locale is already applied locally */ })
    }
  }

  return {
    currentLocale: locale,
    availableLocales,
    setLocale,
  }
}
