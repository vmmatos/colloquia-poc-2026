export interface AppNotification {
  id: number
  type: 'message' | 'mention' | 'agent'
  title: string
  body: string
  time: string
  read: boolean
  channelId?: string
}

const notifications = ref<AppNotification[]>([])

export function useNotifications() {
  const unreadCount = computed(() =>
    new Set(notifications.value.filter(n => !n.read && n.channelId).map(n => n.channelId)).size
  )

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
