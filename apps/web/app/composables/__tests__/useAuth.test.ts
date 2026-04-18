import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useAuth } from '../useAuth'

describe('useAuth', () => {
  beforeEach(() => {
    vi.stubGlobal('$fetch', vi.fn())
    // Reset auth state between tests
    const { auth } = useAuth()
    auth.value = {
      user_id: null,
      access_token: null,
      expires_at: null,
    }
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  // State initialization
  it('initializes with empty auth state', () => {
    const { auth } = useAuth()
    expect(auth.value.user_id).toBeNull()
    expect(auth.value.access_token).toBeNull()
    expect(auth.value.expires_at).toBeNull()
  })

  // Computed properties
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

  // Login
  it('login success populates auth state', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({
      user_id: 'user123',
      access_token: 'token123',
      expires_at: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
    })

    const { auth, login } = useAuth()
    await login('test@example.com', 'password')

    expect(mockFetch).toHaveBeenCalledWith('/api/auth/login', {
      method: 'POST',
      body: { email: 'test@example.com', password: 'password' },
    })
    expect(auth.value.user_id).toBe('user123')
    expect(auth.value.access_token).toBe('token123')
  })

  it('login failure leaves auth state empty', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockRejectedValueOnce(new Error('Invalid credentials'))

    const { auth, login } = useAuth()
    await expect(login('test@example.com', 'wrong')).rejects.toThrow()

    expect(auth.value.user_id).toBeNull()
    expect(auth.value.access_token).toBeNull()
  })

  // Register
  it('register success populates auth state', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({
      user_id: 'user456',
      access_token: 'token456',
      expires_at: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
    })

    const { auth, register } = useAuth()
    await register('new@example.com', 'password')

    expect(mockFetch).toHaveBeenCalledWith('/api/auth/register', {
      method: 'POST',
      body: { email: 'new@example.com', password: 'password' },
    })
    expect(auth.value.user_id).toBe('user456')
    expect(auth.value.access_token).toBe('token456')
  })

  // Logout
  it('logout clears auth state and calls endpoint', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({})

    const { auth, logout } = useAuth()
    auth.value = {
      user_id: 'user789',
      access_token: 'token789',
      expires_at: new Date().toISOString(),
    }

    await logout()

    expect(mockFetch).toHaveBeenCalledWith('/api/auth/logout', {
      method: 'POST',
      body: { access_token: 'token789' },
    })
    expect(auth.value.user_id).toBeNull()
    expect(auth.value.access_token).toBeNull()
  })

  // Refresh Token
  it('refreshToken updates access token', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({
      user_id: 'user123',
      access_token: 'newToken',
      expires_at: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
    })

    const { auth, refreshToken } = useAuth()
    auth.value.user_id = 'user123'
    auth.value.access_token = 'oldToken'

    await refreshToken()

    expect(mockFetch).toHaveBeenCalledWith('/api/auth/refresh', {
      method: 'POST',
    })
    expect(auth.value.access_token).toBe('newToken')
  })

  // Profile methods
  it('getProfile calls correct endpoint with auth header', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({ name: 'John', email: 'john@example.com' })

    const { auth, getProfile } = useAuth()
    auth.value.access_token = 'token123'

    await getProfile()

    expect(mockFetch).toHaveBeenCalledWith('/api/users/me', {
      headers: { Authorization: 'Bearer token123' },
    })
  })

  it('validateToken calls correct endpoint with auth header', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({ valid: true })

    const { auth, validateToken } = useAuth()
    auth.value.access_token = 'token123'

    await validateToken()

    expect(mockFetch).toHaveBeenCalledWith('/api/v1/auth/me', {
      headers: { Authorization: 'Bearer token123' },
    })
  })
})
