export interface Message {
  id: string
  channel_id: string
  user_id: string
  content: string
  created_at: number // Unix timestamp (segundos)
}

export function useMessaging() {
  const { authFetch } = useAuthFetch()

  async function fetchMessages(
    channelId: string,
    opts: { beforeId?: string; limit?: number } = {},
  ): Promise<Message[]> {
    const query: Record<string, string | number> = {
      channel_id: channelId,
      limit: opts.limit ?? 50,
    }
    if (opts.beforeId) query.before_id = opts.beforeId

    const result = await authFetch<Message[]>('/api/messages', { query })
    return result.reverse()
  }

  async function sendMessage(channelId: string, content: string): Promise<Message> {
    return await authFetch<Message>('/api/messages', {
      method: 'POST',
      body: { channel_id: channelId, content },
    })
  }

  return { fetchMessages, sendMessage }
}
