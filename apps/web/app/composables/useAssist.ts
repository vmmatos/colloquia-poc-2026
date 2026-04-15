export function useAssist() {
  const { authFetch } = useAuthFetch()

  const suggestions = ref<string[]>([])
  const isLoading = ref(false)
  let debounceTimer: ReturnType<typeof setTimeout> | null = null
  let abortController: AbortController | null = null

  function clearSuggestions() {
    if (debounceTimer !== null) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    if (abortController) {
      abortController.abort()
      abortController = null
    }
    suggestions.value = []
    isLoading.value = false
  }

  function debouncedFetch(channelId: string, inputText: string) {
    if (debounceTimer !== null) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }

    if (inputText.length <= 10) {
      suggestions.value = []
      isLoading.value = false
      return
    }

    debounceTimer = setTimeout(async () => {
      debounceTimer = null

      // Cancel any previous in-flight request
      if (abortController) {
        abortController.abort()
      }
      abortController = new AbortController()

      // Timeout: abort after 10s so we never block the UI waiting for a slow backend.
      const timeoutId = setTimeout(() => abortController?.abort(), 10_000)

      isLoading.value = true
      try {
        const data = await authFetch<{ suggestions: string[] }>('/api/assist/suggestions', {
          method: 'POST',
          body: { channel_id: channelId, current_input: inputText, message_limit: 10 },
          signal: abortController.signal,
        })
        suggestions.value = (data.suggestions ?? []).slice(0, 3)
      }
      catch (e: unknown) {
        // Ignore aborted requests — a newer fetch is already in-flight or timed out
        if (e instanceof Error && e.name === 'AbortError') return
        suggestions.value = []
      }
      finally {
        clearTimeout(timeoutId)
        isLoading.value = false
        abortController = null
      }
    }, 500)
  }

  onUnmounted(() => {
    if (debounceTimer !== null) clearTimeout(debounceTimer)
    if (abortController) abortController.abort()
  })

  return { suggestions, isLoading, debouncedFetch, clearSuggestions }
}
