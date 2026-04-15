export function useAuthFetch() {
  const { auth, refreshToken } = useAuth()

  async function authFetch<T>(url: string, opts: Parameters<typeof $fetch>[1] = {}): Promise<T> {
    const doFetch = () =>
      $fetch<T>(url, {
        ...opts,
        headers: {
          ...opts.headers,
          Authorization: `Bearer ${auth.value.access_token}`,
        },
      }) as Promise<T>

    try {
      return await doFetch()
    } catch (e: any) {
      if (e?.response?.status === 401) {
        try {
          await refreshToken()
          return await doFetch()
        } catch {
          auth.value = { user_id: null, access_token: null, expires_at: null }
          throw e
        }
      }
      throw e
    }
  }

  return { authFetch }
}
