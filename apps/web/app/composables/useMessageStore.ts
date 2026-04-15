import type { Message } from '~/composables/useMessaging'

export interface DisplayMessage {
  id: string
  userId: string
  author: string
  text: string
  time: string
  isAgent?: boolean
}

function formatTime(unixSeconds: number): string {
  return new Date(unixSeconds * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export function messageToDisplay(m: Message, resolveUser: (id: string) => string): DisplayMessage {
  return {
    id: m.id,
    userId: m.user_id,
    author: resolveUser(m.user_id),
    text: m.content,
    time: formatTime(m.created_at),
  }
}

export function useMessageStore() {
  const store = useState<Record<string, DisplayMessage[]>>('messageStore', () => ({}))

  function get(channelId: string): DisplayMessage[] {
    return store.value[channelId] ?? []
  }

  // Replace a channel's messages, preserving any live events that arrived
  // during the history fetch and aren't present in the new list.
  function setHistory(channelId: string, history: DisplayMessage[]) {
    const existing = store.value[channelId] ?? []
    const historyIds = new Set(history.map(m => m.id))
    const extras = existing.filter(m => !historyIds.has(m.id))
    store.value = { ...store.value, [channelId]: [...history, ...extras] }
  }

  function append(channelId: string, message: DisplayMessage) {
    const existing = store.value[channelId] ?? []
    if (existing.some(m => m.id === message.id)) return // dedup by id
    store.value = { ...store.value, [channelId]: [...existing, message] }
  }

  function clearChannel(channelId: string) {
    if (!(channelId in store.value)) return
    const next = { ...store.value }
    delete next[channelId]
    store.value = next
  }

  return { store, get, setHistory, append, clearChannel }
}
