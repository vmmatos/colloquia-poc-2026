interface UserApiResponse {
  user_id: string
  name: string
  email: string
}

// Module-level cache — partilhado por todas as instâncias do composable (como useNotifications)
const cache = reactive<Record<string, string>>({}) // userId → displayName
const pending = new Set<string>()

export function useUsersCache() {
  const { auth } = useAuth()

  async function loadUser(userId: string): Promise<void> {
    if (cache[userId] !== undefined || pending.has(userId)) return
    pending.add(userId)
    try {
      const p = await $fetch<UserApiResponse>(`/api/users/${userId}`)
      cache[userId] = p.name || p.email || userId.slice(0, 8)
    } catch {
      cache[userId] = userId.slice(0, 8)
    } finally {
      pending.delete(userId)
    }
  }

  async function prefetchUsers(userIds: string[]): Promise<void> {
    const missing = userIds.filter(id => id !== auth.value.user_id && cache[id] === undefined)
    await Promise.all(missing.map(loadUser))
  }

  function resolveUser(userId: string): string {
    if (userId === auth.value.user_id) return 'Tu'
    if (cache[userId] !== undefined) return cache[userId]
    loadUser(userId) // background load para renderizações futuras
    return userId.slice(0, 8)
  }

  return { resolveUser, prefetchUsers, loadUser }
}
