import { describe, it, expect } from 'vitest'
import { useAuth } from '../useAuth'

describe('useAuth', () => {
  it('initializes with empty auth state', () => {
    const { auth } = useAuth()
    expect(auth.value.user_id).toBeNull()
    expect(auth.value.access_token).toBeNull()
    expect(auth.value.expires_at).toBeNull()
  })

  it('isAuthenticated computed is false initially', () => {
    const { isAuthenticated } = useAuth()
    expect(isAuthenticated.value).toBe(false)
  })

  it('isAuthenticated computed is true when access_token exists', () => {
    const { auth, isAuthenticated } = useAuth()
    auth.value.access_token = 'token123'
    expect(isAuthenticated.value).toBe(true)
  })

  it('tokenExpiresIn returns null when no expiration', () => {
    const { auth, tokenExpiresIn } = useAuth()
    auth.value.expires_at = null
    expect(tokenExpiresIn.value).toBeNull()
  })

  it('tokenExpiresIn shows "expired" for past time', () => {
    const { auth, tokenExpiresIn } = useAuth()
    auth.value.expires_at = new Date(Date.now() - 1000).toISOString()
    expect(tokenExpiresIn.value).toBe('expired')
  })

  it('tokenExpiresIn shows seconds remaining', () => {
    const { auth, tokenExpiresIn } = useAuth()
    auth.value.expires_at = new Date(Date.now() + 30 * 1000).toISOString()
    expect(tokenExpiresIn.value).toMatch(/\d+s/)
  })

  it('tokenExpiresIn shows minutes and seconds remaining', () => {
    const { auth, tokenExpiresIn } = useAuth()
    auth.value.expires_at = new Date(Date.now() + 5 * 60 * 1000).toISOString()
    expect(tokenExpiresIn.value).toMatch(/[4-5]m \d+s/)
  })
})
