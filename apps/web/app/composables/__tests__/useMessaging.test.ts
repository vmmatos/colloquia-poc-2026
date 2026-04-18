import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { useMessaging } from '../useMessaging'

// Mock useAuthFetch
const mockAuthFetch = vi.fn()
mockNuxtImport('useAuthFetch', () => () => ({
  authFetch: mockAuthFetch,
}))

describe('useMessaging', () => {
  beforeEach(() => {
    mockAuthFetch.mockReset()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('fetchMessages calls correct endpoint', async () => {
    mockAuthFetch.mockResolvedValueOnce([])

    const { fetchMessages } = useMessaging()
    await fetchMessages('ch1')

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      query: { channel_id: 'ch1', limit: 50 },
    })
  })

  it('fetchMessages reverses result (oldest first)', async () => {
    const messages = [
      { id: '3', channel_id: 'ch1', user_id: 'u1', content: 'Third', created_at: 3000 },
      { id: '2', channel_id: 'ch1', user_id: 'u1', content: 'Second', created_at: 2000 },
      { id: '1', channel_id: 'ch1', user_id: 'u1', content: 'First', created_at: 1000 },
    ]
    mockAuthFetch.mockResolvedValueOnce(messages)

    const { fetchMessages } = useMessaging()
    const result = await fetchMessages('ch1')

    expect(result[0].id).toBe('1')
    expect(result[1].id).toBe('2')
    expect(result[2].id).toBe('3')
  })

  it('fetchMessages respects custom limit', async () => {
    mockAuthFetch.mockResolvedValueOnce([])

    const { fetchMessages } = useMessaging()
    await fetchMessages('ch1', { limit: 100 })

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      query: { channel_id: 'ch1', limit: 100 },
    })
  })

  it('fetchMessages includes beforeId in query', async () => {
    mockAuthFetch.mockResolvedValueOnce([])

    const { fetchMessages } = useMessaging()
    await fetchMessages('ch1', { beforeId: 'msg123' })

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      query: { channel_id: 'ch1', limit: 50, before_id: 'msg123' },
    })
  })

  it('fetchMessages ignores undefined beforeId', async () => {
    mockAuthFetch.mockResolvedValueOnce([])

    const { fetchMessages } = useMessaging()
    await fetchMessages('ch1', { beforeId: undefined })

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      query: { channel_id: 'ch1', limit: 50 },
    })
  })

  it('sendMessage posts to correct endpoint', async () => {
    const msg = {
      id: 'msg1',
      channel_id: 'ch1',
      user_id: 'u1',
      content: 'Hello',
      created_at: 1000,
    }
    mockAuthFetch.mockResolvedValueOnce(msg)

    const { sendMessage } = useMessaging()
    const result = await sendMessage('ch1', 'Hello')

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      method: 'POST',
      body: { channel_id: 'ch1', content: 'Hello' },
    })
    expect(result).toEqual(msg)
  })

  it('sendMessage returns the created message', async () => {
    const msg = {
      id: 'msg1',
      channel_id: 'ch1',
      user_id: 'u1',
      content: 'Test',
      created_at: 1000,
    }
    mockAuthFetch.mockResolvedValueOnce(msg)

    const { sendMessage } = useMessaging()
    const result = await sendMessage('ch1', 'Test')

    expect(result).toEqual(msg)
  })

  it('sendMessage sends any content (even empty)', async () => {
    mockAuthFetch.mockResolvedValueOnce({})

    const { sendMessage } = useMessaging()
    await sendMessage('ch1', '')

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      method: 'POST',
      body: { channel_id: 'ch1', content: '' },
    })
  })

  it('fetchMessages with both limit and beforeId', async () => {
    mockAuthFetch.mockResolvedValueOnce([])

    const { fetchMessages } = useMessaging()
    await fetchMessages('ch1', { limit: 25, beforeId: 'msg999' })

    expect(mockAuthFetch).toHaveBeenCalledWith('/api/messages', {
      query: { channel_id: 'ch1', limit: 25, before_id: 'msg999' },
    })
  })
})