interface PresenceEvent {
  user_id: string
  online: boolean
  last_seen: number
}

export function usePresence(options?: { EventSourceCtor?: typeof EventSource }) {
  const { auth } = useAuth()
  const presenceMap = useState<Record<string, boolean>>('presenceMap', () => ({}))

  let source: EventSource | null = null
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null
  let retryTimer: ReturnType<typeof setTimeout> | null = null
  let retryCount = 0

  const EventSourceCtor = options?.EventSourceCtor ?? (typeof EventSource !== 'undefined' ? EventSource : null)

  function openPresenceStream() {
    if (source || !auth.value.access_token || !EventSourceCtor) return

    const params = new URLSearchParams({ token: auth.value.access_token })
    source = new EventSourceCtor(`/api/v1/users/presence/stream?${params}`)

    source.addEventListener('message', (e: MessageEvent) => {
      try {
        const evt: PresenceEvent = JSON.parse(e.data)
        presenceMap.value[evt.user_id] = evt.online
        retryCount = 0
      } catch {
        // Malformed event — ignore
      }
    })

    source.addEventListener('error', () => {
      source?.close()
      source = null
      const delay = Math.min(2000 * Math.pow(2, retryCount), 30_000)
      retryCount++
      retryTimer = setTimeout(openPresenceStream, delay)
    })
  }

  function closePresenceStream() {
    if (retryTimer) {
      clearTimeout(retryTimer)
      retryTimer = null
    }
    source?.close()
    source = null
  }

  function startHeartbeat() {
    if (heartbeatTimer) return
    sendHeartbeat()
    heartbeatTimer = setInterval(sendHeartbeat, 10_000)
  }

  function stopHeartbeat() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }

  async function sendHeartbeat() {
    if (!auth.value.access_token) return
    try {
      await $fetch('/api/v1/users/heartbeat', {
        method: 'POST',
        headers: { Authorization: `Bearer ${auth.value.access_token}` },
      })
    } catch {
      // Best-effort — server marks offline after 90s with no heartbeat
    }
  }

  function isOnline(userId: string): boolean {
    return presenceMap.value[userId] ?? false
  }

  function init() {
    openPresenceStream()
    startHeartbeat()
  }

  function destroy() {
    closePresenceStream()
    stopHeartbeat()
  }

  return { presenceMap, isOnline, init, destroy }
}
