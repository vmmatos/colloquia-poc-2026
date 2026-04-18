import { describe, it, expect } from 'vitest'
import { useNotifications } from '../useNotifications'

describe('useNotifications', () => {
  it('addNotification with read=false by default', () => {
    const { addNotification, notifications } = useNotifications()

    addNotification({
      type: 'message',
      title: 'Test',
      body: 'Body',
      time: '10:00',
    })

    expect(notifications.value[notifications.value.length - 1]).toMatchObject({
      type: 'message',
      title: 'Test',
      read: false,
    })
  })

  it('markAllRead marks all notifications as read', () => {
    const { notifications, addNotification, markAllRead } = useNotifications()
    notifications.value = [] // Reset for this test

    addNotification({ type: 'message', title: 'A', body: 'A', time: '10:00' })
    addNotification({ type: 'message', title: 'B', body: 'B', time: '10:01' })

    markAllRead()

    expect(notifications.value.every(n => n.read)).toBe(true)
  })

  it('markRead marks specific notification by id as read', () => {
    const { notifications, addNotification, markRead } = useNotifications()
    notifications.value = []

    addNotification({ type: 'message', title: 'A', body: 'A', time: '10:00' })
    const firstId = notifications.value[0].id

    addNotification({ type: 'message', title: 'B', body: 'B', time: '10:01' })

    markRead(firstId)

    expect(notifications.value.find(n => n.id === firstId)?.read).toBe(true)
  })

  it('unreadCount computes unique channels with unread', () => {
    const { notifications, addNotification, unreadCount } = useNotifications()
    notifications.value = []

    addNotification({
      type: 'message',
      title: 'A',
      body: 'A',
      time: '10:00',
      channelId: 'ch1',
    })
    addNotification({
      type: 'message',
      title: 'B',
      body: 'B',
      time: '10:01',
      channelId: 'ch1',
    })
    addNotification({
      type: 'message',
      title: 'C',
      body: 'C',
      time: '10:02',
      channelId: 'ch2',
    })

    expect(unreadCount.value).toBe(2)
  })

  it('unreadCount excludes notifications without channelId', () => {
    const { notifications, addNotification, unreadCount } = useNotifications()
    notifications.value = []

    addNotification({ type: 'mention', title: 'A', body: 'A', time: '10:00' })
    addNotification({
      type: 'message',
      title: 'B',
      body: 'B',
      time: '10:01',
      channelId: 'ch1',
    })

    expect(unreadCount.value).toBe(1)
  })

  it('markChannelRead marks channel notifications as read', () => {
    const { notifications, addNotification, markChannelRead } = useNotifications()
    notifications.value = []

    addNotification({
      type: 'message',
      title: 'A',
      body: 'A',
      time: '10:00',
      channelId: 'ch1',
    })
    addNotification({
      type: 'message',
      title: 'B',
      body: 'B',
      time: '10:01',
      channelId: 'ch2',
    })

    markChannelRead('ch1')

    const ch1 = notifications.value.find(n => n.channelId === 'ch1')
    const ch2 = notifications.value.find(n => n.channelId === 'ch2')

    expect(ch1?.read).toBe(true)
    expect(ch2?.read).toBe(false)
  })
})
