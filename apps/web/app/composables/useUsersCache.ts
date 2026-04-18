interface UserApiResponse {
  user_id: string
  name: string
  email: string
}

export function useUsersCache() {
  const { auth } = useAuth()
  const cache = useState<Record<string, string>>('usersCache', () => ({}))
  const pending = useState<Set<string>>('usersCachePending', () => new Set())

  async function loadUser(userId: string): Promise<void> {
    if (cache.value[userId] !== undefined || pending.value.has(userId)) return
    pending.value.add(userId)
    try {
      const p = await $fetch<UserApiResponse>(`/api/users/${userId}`)
      cache.value[userId] = p.name || p.email || userId.slice(0, 8)
    } catch {
      cache.value[userId] = userId.slice(0, 8)
    } finally {
      pending.value.delete(userId)
    }
  }

  async function prefetchUsers(userIds: string[]): Promise<void> {
    const missing = userIds.filter(id => id !== auth.value.user_id && cache.value[id] === undefined)
    await Promise.all(missing.map(loadUser))
  }

  function resolveUser(userId: string): string {
    if (userId === auth.value.user_id) return 'Tu'
    if (cache.value[userId] !== undefined) return cache.value[userId]
    loadUser(userId)
    return userId.slice(0, 8)
  }

  return { resolveUser, prefetchUsers, loadUser }
}
