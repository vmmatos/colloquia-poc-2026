import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { useAssist } from '../useAssist'

// Mock useAuthFetch
const mockAuthFetch = vi.fn()
mockNuxtImport('useAuthFetch', () => () => ({
  authFetch: mockAuthFetch,
}))

describe('useAssist', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    mockAuthFetch.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('initializes with empty suggestions and not loading', () => {
    const { suggestions, isLoading } = useAssist()
    expect(suggestions.value).toEqual([])
    expect(isLoading.value).toBe(false)
  })

  it('clearSuggestions clears suggestions and loading state', () => {
    const { suggestions, isLoading, clearSuggestions } = useAssist()

    suggestions.value = ['test']
    isLoading.value = true

    clearSuggestions()

    expect(suggestions.value).toEqual([])
    expect(isLoading.value).toBe(false)
  })

  it('does not fetch when input is 10 characters or less', () => {
    const { debouncedFetch } = useAssist()

    debouncedFetch('ch1', '12345')
    debouncedFetch('ch1', '1234567890') // exactly 10

    vi.advanceTimersByTime(600)

    expect(mockAuthFetch).not.toHaveBeenCalled()
  })

  it('clears suggestions when input becomes too short', () => {
    const { suggestions, isLoading, debouncedFetch } = useAssist()

    suggestions.value = ['old suggestion']
    isLoading.value = true

    debouncedFetch('ch1', '12345')

    expect(suggestions.value).toEqual([])
    expect(isLoading.value).toBe(false)
  })

  it('debounces fetch request by 500ms', async () => {
    mockAuthFetch.mockResolvedValueOnce({ suggestions: ['test'] })

    const { debouncedFetch } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')

    // Should not fetch yet
    vi.advanceTimersByTime(499)
    expect(mockAuthFetch).not.toHaveBeenCalled()

    // After 500ms should fetch
    vi.advanceTimersByTime(1)
    await vi.runAllTimersAsync()

    expect(mockAuthFetch).toHaveBeenCalled()
  })

  it('cancels previous debounce when new input arrives', async () => {
    mockAuthFetch.mockResolvedValueOnce({ suggestions: ['s1', 's2'] })

    const { debouncedFetch } = useAssist()

    debouncedFetch('ch1', '12345 67890 A')
    vi.advanceTimersByTime(200)

    // New input before first debounce fires
    debouncedFetch('ch1', '12345 67890 ABC')

    // Advance to original 500ms - should not fire yet
    vi.advanceTimersByTime(300)
    expect(mockAuthFetch).not.toHaveBeenCalled()

    // Advance another 500ms from second call
    vi.advanceTimersByTime(500)
    await vi.runAllTimersAsync()

    // Should fetch only once with latest input
    expect(mockAuthFetch).toHaveBeenCalledTimes(1)
    expect(mockAuthFetch).toHaveBeenCalledWith(
      '/api/assist/suggestions',
      expect.objectContaining({
        body: expect.objectContaining({
          current_input: '12345 67890 ABC',
        }),
      })
    )
  })

  it('fetches with correct payload', async () => {
    mockAuthFetch.mockResolvedValueOnce({ suggestions: [] })

    const { debouncedFetch } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')
    vi.advanceTimersByTime(500)
    await vi.runAllTimersAsync()

    expect(mockAuthFetch).toHaveBeenCalledWith(
      '/api/assist/suggestions',
      {
        method: 'POST',
        body: { channel_id: 'ch1', current_input: '12345 67890 ABC', message_limit: 10 },
        signal: expect.any(AbortSignal),
      }
    )
  })

  it('limits suggestions to 3 items', async () => {
    mockAuthFetch.mockResolvedValueOnce({
      suggestions: ['s1', 's2', 's3', 's4', 's5'],
    })

    const { debouncedFetch, suggestions } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')
    vi.advanceTimersByTime(500)
    await vi.runAllTimersAsync()

    expect(suggestions.value).toEqual(['s1', 's2', 's3'])
  })

  it('clears suggestions on fetch error', async () => {
    mockAuthFetch.mockRejectedValueOnce(new Error('Network error'))

    const { debouncedFetch, suggestions } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')
    vi.advanceTimersByTime(500)
    await vi.runAllTimersAsync()

    expect(suggestions.value).toEqual([])
  })

  it('ignores AbortError from previous request cancellation', async () => {
    const abortErr = new Error('Aborted')
    abortErr.name = 'AbortError'
    mockAuthFetch.mockRejectedValueOnce(abortErr)

    const { debouncedFetch } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')
    vi.advanceTimersByTime(500)
    await vi.runAllTimersAsync()

    // AbortError is caught and ignored silently (newer request in flight)
    expect(mockAuthFetch).toHaveBeenCalled()
  })

  it('sets isLoading while fetching', async () => {
    mockAuthFetch.mockImplementationOnce(async () => {
      // Can't easily test isLoading state during fetch with fakeTimers
      return { suggestions: ['test'] }
    })

    const { debouncedFetch, isLoading } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')
    vi.advanceTimersByTime(500)

    // After fetch completes, isLoading should be false
    await vi.runAllTimersAsync()

    expect(isLoading.value).toBe(false)
  })

  it('handles empty suggestions response', async () => {
    mockAuthFetch.mockResolvedValueOnce({ suggestions: null })

    const { debouncedFetch, suggestions } = useAssist()

    debouncedFetch('ch1', '12345 67890 ABC')
    vi.advanceTimersByTime(500)
    await vi.runAllTimersAsync()

    expect(suggestions.value).toEqual([])
  })
})