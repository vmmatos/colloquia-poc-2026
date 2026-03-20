export default defineNuxtPlugin(async () => {
  const authReady = useState<boolean>('authReady', () => false)
  const { isAuthenticated, refreshToken } = useAuth()
  if (!isAuthenticated.value) {
    try { await refreshToken() } catch { /* no valid cookie, middleware will redirect */ }
  }
  authReady.value = true
})
