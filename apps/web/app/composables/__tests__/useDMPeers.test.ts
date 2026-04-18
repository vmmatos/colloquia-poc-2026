import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import { useDMPeers } from '../useDMPeers'

// Mock useAuth
const mockAuth = ref({ user_id: 'me' })
mockNuxtImport('useAuth', () => () => ({
  auth: mockAuth,
}))

describe('useDMPeers', () => {
  beforeEach(() => {
    mockAuth.value = { user_id: 'me' }
    const { peers } = useDMPeers()
    peers.value = {}
  })

  it('setPeer stores other user ID for a channel', () => {
    const { setPeer, peers } = useDMPeers()

    setPeer('ch1', ['me', 'other'])

    expect(peers.value['ch1']).toBe('other')
  })

  it('setPeer handles reverse order of IDs', () => {
    const { setPeer, peers } = useDMPeers()

    setPeer('ch1', ['other', 'me'])

    expect(peers.value['ch1']).toBe('other')
  })

  it('setPeer does not store when both IDs are current user', () => {
    const { setPeer, peers } = useDMPeers()

    setPeer('ch1', ['me', 'me'])

    expect(peers.value['ch1']).toBeUndefined()
  })

  it('setPeer does not store when only current user in list', () => {
    const { setPeer, peers } = useDMPeers()

    setPeer('ch1', ['me'])

    expect(peers.value['ch1']).toBeUndefined()
  })

  it('getPeer returns stored user ID', () => {
    const { setPeer, getPeer } = useDMPeers()

    setPeer('ch1', ['me', 'alice'])
    const result = getPeer('ch1')

    expect(result).toBe('alice')
  })

  it('getPeer returns null for unknown channel', () => {
    const { getPeer } = useDMPeers()

    const result = getPeer('unknown')

    expect(result).toBeNull()
  })

  it('multiple peers are stored independently', () => {
    const { setPeer, getPeer } = useDMPeers()

    setPeer('ch1', ['me', 'alice'])
    setPeer('ch2', ['me', 'bob'])

    expect(getPeer('ch1')).toBe('alice')
    expect(getPeer('ch2')).toBe('bob')
  })

  it('setPeer overwrites previous peer for same channel', () => {
    const { setPeer, getPeer } = useDMPeers()

    setPeer('ch1', ['me', 'alice'])
    expect(getPeer('ch1')).toBe('alice')

    setPeer('ch1', ['me', 'bob'])
    expect(getPeer('ch1')).toBe('bob')
  })

  it('peers is reactive reference', () => {
    const { peers } = useDMPeers()

    peers.value = { ch1: 'alice', ch2: 'bob' }

    expect(peers.value['ch1']).toBe('alice')
    expect(peers.value['ch2']).toBe('bob')
  })
})