export interface Message {
  id: string
  channel_id: string
  user_id: string
  content: string
  created_at: number // Unix timestamp (segundos)
}

export function useMessaging() {
  const { auth } = useAuth()

  function authHeaders() {
    return { Authorization: `Bearer ${auth.value.access_token}` }
  }

  async function fetchMessages(
    channelId: string,
    opts: { beforeId?: string; limit?: number } = {},
  ): Promise<Message[]> {
    const query: Record<string, string | number> = {
      channel_id: channelId,
      limit: opts.limit ?? 50,
    }
    if (opts.beforeId) query.before_id = opts.beforeId

    const result = await $fetch<Message[]>('/api/messages', {
      query,
      headers: authHeaders(),
    })
    // Backend retorna DESC (mais recente primeiro); reverter para exibir cronologicamente
    return result.reverse()
  }

  async function sendMessage(channelId: string, content: string): Promise<Message> {
    return await $fetch<Message>('/api/messages', {
      method: 'POST',
      body: { channel_id: channelId, content },
      headers: authHeaders(),
    })
  }

  return { fetchMessages, sendMessage }
}
