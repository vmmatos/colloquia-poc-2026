import { describe, it, expect } from 'vitest'
import { useMessageStore, type DisplayMessage } from '../useMessageStore'

describe('useMessageStore', () => {
  it('get returns empty array for unknown channel', () => {
    const { get } = useMessageStore()
    expect(get('unknown')).toEqual([])
  })

  it('append adds message to channel', () => {
    const { append, get } = useMessageStore()
    const msg: DisplayMessage = {
      id: '1',
      userId: 'user1',
      author: 'Alice',
      text: 'Hello',
      time: '10:00',
    }

    append('channel1', msg)
    const result = get('channel1')

    expect(result).toContainEqual(msg)
  })

  it('append deduplicates by message id', () => {
    const { append, get } = useMessageStore()
    const msg: DisplayMessage = {
      id: 'dup-1',
      userId: 'user1',
      author: 'Alice',
      text: 'Hello',
      time: '10:00',
    }

    append('channel1', msg)
    const len1 = get('channel1').length
    append('channel1', msg)
    const len2 = get('channel1').length

    expect(len2).toBe(len1)
  })

  it('setHistory merges with existing messages', () => {
    const { append, setHistory, get } = useMessageStore()

    const liveMsg: DisplayMessage = {
      id: 'live-1',
      userId: 'user1',
      author: 'Alice',
      text: 'Live',
      time: '10:00',
    }

    append('channel1', liveMsg)

    const history: DisplayMessage[] = [
      {
        id: 'hist-1',
        userId: 'user1',
        author: 'Alice',
        text: 'History',
        time: '09:00',
      },
    ]

    setHistory('channel1', history)
    const messages = get('channel1')

    expect(messages.map(m => m.id)).toContain('hist-1')
    expect(messages.map(m => m.id)).toContain('live-1')
  })

  it('clearChannel is no-op for unknown channel', () => {
    const { clearChannel } = useMessageStore()
    expect(() => clearChannel('unknown')).not.toThrow()
  })

  it('append adds message with content properties', () => {
    const { append, get } = useMessageStore()
    const msg: DisplayMessage = {
      id: 'msg-1',
      userId: 'u1',
      author: 'Author',
      text: 'Content',
      time: '10:00',
      isAgent: true,
    }

    append('ch1', msg)
    const stored = get('ch1')[get('ch1').length - 1]

    expect(stored.userId).toBe('u1')
    expect(stored.text).toBe('Content')
    expect(stored.isAgent).toBe(true)
  })
})
