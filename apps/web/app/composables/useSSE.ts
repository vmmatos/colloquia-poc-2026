export interface SseEvent {
  id: string
  channel_id: string
  user_id: string
  content: string
  created_at: number // Unix timestamp (segundos)
}

export function useSSE(options: {
  activeChannelId: Ref<string | null>
  onMessage: (event: SseEvent) => void
}) {
  const { auth } = useAuth()

  // Map simples (não reactivo — gestão interna de conexões)
  const connections = new Map<
    string,
    {
      source: EventSource
      retryCount: number
      retryTimer: ReturnType<typeof setTimeout> | null
    }
  >()

  function openConnection(channelId: string) {
    if (connections.has(channelId) || !auth.value.access_token) return

    const token = auth.value.access_token
    const url = `/api/v1/messages/stream?channel_id=${channelId}&token=${token}`
    const source = new EventSource(url)
    const entry = {
      source,
      retryCount: 0,
      retryTimer: null as ReturnType<typeof setTimeout> | null,
    }
    connections.set(channelId, entry)

    source.addEventListener('message', (e: MessageEvent) => {
      try {
        const payload: SseEvent = JSON.parse(e.data)
        entry.retryCount = 0
        options.onMessage(payload)
      } catch {
        // JSON inválido — ignorar
      }
    })

    source.addEventListener('error', () => {
      source.close()
      connections.delete(channelId)
      const delay = Math.min(2000 * Math.pow(2, entry.retryCount), 30_000)
      entry.retryCount++
      entry.retryTimer = setTimeout(() => openConnection(channelId), delay)
    })
  }

  function closeConnection(channelId: string) {
    const entry = connections.get(channelId)
    if (!entry) return
    if (entry.retryTimer) clearTimeout(entry.retryTimer)
    entry.source.close()
    connections.delete(channelId)
  }

  function subscribeToChannels(channelIds: string[]) {
    for (const id of connections.keys()) {
      if (!channelIds.includes(id)) closeConnection(id)
    }
    for (const id of channelIds) openConnection(id)
  }

  function closeAll() {
    for (const id of [...connections.keys()]) closeConnection(id)
  }

  return { subscribeToChannels, closeAll }
}
