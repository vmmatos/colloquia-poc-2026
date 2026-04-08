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

  let source: EventSource | null = null
  let retryCount = 0
  let retryTimer: ReturnType<typeof setTimeout> | null = null
  let subscribedChannels: string[] = []

  function openConnection(channelIds: string[]) {
    if (source || channelIds.length === 0 || !auth.value.access_token) return

    const params = new URLSearchParams()
    channelIds.forEach(id => params.append('channel_id', id))
    params.set('token', auth.value.access_token)

    source = new EventSource(`/api/v1/messages/stream?${params.toString()}`)

    source.addEventListener('message', (e: MessageEvent) => {
      try {
        const payload: SseEvent = JSON.parse(e.data)
        retryCount = 0
        options.onMessage(payload)
      } catch {
        // JSON inválido — ignorar
      }
    })

    source.addEventListener('error', () => {
      source?.close()
      source = null
      const delay = Math.min(2000 * Math.pow(2, retryCount), 30_000)
      retryCount++
      retryTimer = setTimeout(() => openConnection(subscribedChannels), delay)
    })
  }

  function closeConnection() {
    if (retryTimer) {
      clearTimeout(retryTimer)
      retryTimer = null
    }
    source?.close()
    source = null
  }

  function subscribeToChannels(channelIds: string[]) {
    const prev = subscribedChannels.slice().sort().join(',')
    const next = channelIds.slice().sort().join(',')
    subscribedChannels = channelIds

    if (prev === next) return // lista não mudou, manter conexão

    closeConnection()
    retryCount = 0
    openConnection(channelIds)
  }

  function closeAll() {
    subscribedChannels = []
    closeConnection()
  }

  return { subscribeToChannels, closeAll }
}
