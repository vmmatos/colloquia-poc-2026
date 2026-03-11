export interface AppNotification {
  id: number
  type: 'message' | 'mention' | 'agent'
  title: string
  body: string
  time: string
  read: boolean
}

const notifications = ref<AppNotification[]>([
  {
    id: 1,
    type: 'mention',
    title: 'Alice mencionou-te em #general',
    body: 'Ei @you, o que achas desta proposta?',
    time: 'há 2 min',
    read: false,
  },
  {
    id: 2,
    type: 'agent',
    title: 'Agente LLM respondeu',
    body: 'Analisando o contexto da conversa, identifico três pontos principais...',
    time: 'há 5 min',
    read: false,
  },
  {
    id: 3,
    type: 'message',
    title: 'Nova mensagem em #dev',
    body: 'Bob: Alguém reviu o PR #42?',
    time: 'há 12 min',
    read: true,
  },
  {
    id: 4,
    type: 'mention',
    title: 'Charlie mencionou-te em #random',
    body: '@you já viste o novo episódio?',
    time: 'há 1h',
    read: true,
  },
])

export function useNotifications() {
  const unreadCount = computed(() => notifications.value.filter(n => !n.read).length)

  function addNotification(n: Omit<AppNotification, 'id' | 'read'>) {
    notifications.value.unshift({ ...n, id: Date.now(), read: false })
  }

  function markAllRead() {
    notifications.value = notifications.value.map(n => ({ ...n, read: true }))
  }

  function markRead(id: number) {
    const n = notifications.value.find(n => n.id === id)
    if (n) n.read = true
  }

  return { notifications, unreadCount, addNotification, markAllRead, markRead }
}
