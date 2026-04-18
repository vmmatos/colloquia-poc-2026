import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import { useAuthFetch } from '../useAuthFetch'

// Mock useAuth at module level (mockNuxtImport is a macro)
const mockAuth = ref({ access_token: 'token123' })
const mockRefreshToken = vi.fn()

mockNuxtImport('useAuth', () => () => ({
  auth: mockAuth,
  refreshToken: mockRefreshToken,
}))

describe('useAuthFetch', () => {
  beforeEach(() => {
    vi.stubGlobal('$fetch', vi.fn())
    mockAuth.value = { access_token: 'token123' }
    mockRefreshToken.mockReset()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('includes Authorization header in requests', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({ id: 1, name: 'test' })

    const { authFetch } = useAuthFetch()
    await authFetch('/api/users/1')

    expect(mockFetch).toHaveBeenCalledWith(
      '/api/users/1',
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer token123',
        }),
      })
    )
  })

  it('merges custom headers with Authorization', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({})

    const { authFetch } = useAuthFetch()
    await authFetch('/api/test', {
      headers: { 'X-Custom': 'value' },
    })

    expect(mockFetch).toHaveBeenCalledWith(
      '/api/test',
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer token123',
          'X-Custom': 'value',
        }),
      })
    )
  })

  it('returns data on successful request', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    const data = { id: 1, name: 'John' }
    mockFetch.mockResolvedValueOnce(data)

    const { authFetch } = useAuthFetch()
    const result = await authFetch('/api/users/1')

    expect(result).toEqual(data)
    expect(mockFetch).toHaveBeenCalledTimes(1)
  })

  it('retries on 401 after refreshing token', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    const error401 = new Error('Unauthorized')
    ;(error401 as any).response = { status: 401 }

    const data = { id: 1 }
    mockFetch
      .mockRejectedValueOnce(error401) // first call fails
      .mockResolvedValueOnce(data) // second call succeeds

    mockRefreshToken.mockResolvedValueOnce({
      user_id: 'user1',
      access_token: 'newtoken',
      expires_at: null,
    })

    const { authFetch } = useAuthFetch()
    const result = await authFetch('/api/test')

    expect(mockRefreshToken).toHaveBeenCalledTimes(1)
    expect(mockFetch).toHaveBeenCalledTimes(2)
    expect(result).toEqual(data)
  })

  it('clears auth state when refresh fails after 401', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    const error401 = new Error('Unauthorized')
    ;(error401 as any).response = { status: 401 }

    mockFetch.mockRejectedValue(error401)
    mockRefreshToken.mockRejectedValueOnce(new Error('Refresh failed'))

    const { authFetch } = useAuthFetch()

    try {
      await authFetch('/api/test')
    } catch {
      // Expected to throw
    }

    expect(mockAuth.value).toEqual({
      user_id: null,
      access_token: null,
      expires_at: null,
    })
  })

  it('does not retry on non-401 errors', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    const error500 = new Error('Server error')
    ;(error500 as any).response = { status: 500 }

    mockFetch.mockRejectedValueOnce(error500)

    const { authFetch } = useAuthFetch()

    await expect(authFetch('/api/test')).rejects.toThrow('Server error')
    expect(mockRefreshToken).not.toHaveBeenCalled()
    expect(mockFetch).toHaveBeenCalledTimes(1)
  })

  it('propagates options to $fetch call', async () => {
    const mockFetch = vi.mocked(global.$fetch as any)
    mockFetch.mockResolvedValueOnce({})

    const { authFetch } = useAuthFetch()
    await authFetch('/api/test', {
      method: 'POST',
      body: { data: 'test' },
    })

    expect(mockFetch).toHaveBeenCalledWith(
      '/api/test',
      expect.objectContaining({
        method: 'POST',
        body: { data: 'test' },
      })
    )
  })
})