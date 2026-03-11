export function useTokenRefresh() {
  const { auth, refreshToken } = useAuth()
  let timerId: ReturnType<typeof setTimeout> | null = null

  function scheduleRefresh(expiresAt: string | null) {
    if (timerId) {
      clearTimeout(timerId)
      timerId = null
    }
    if (!expiresAt) return

    const msUntilRefresh = new Date(expiresAt).getTime() - Date.now() - 60_000
    if (msUntilRefresh <= 0) return

    timerId = setTimeout(async () => {
      try {
        await refreshToken()
      } catch {
        // token refresh failed — let the auth middleware handle redirect
      }
    }, msUntilRefresh)
  }

  watch(() => auth.value.expires_at, scheduleRefresh, { immediate: true })

  onUnmounted(() => {
    if (timerId) clearTimeout(timerId)
  })
}
