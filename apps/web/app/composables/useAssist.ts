export function useAssist() {
  const { auth } = useAuth()

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

      isLoading.value = true
      try {
        const data = await $fetch<{ suggestions: string[] }>('/api/assist/suggestions', {
          method: 'POST',
          body: { channel_id: channelId, current_input: inputText, message_limit: 10 },
          headers: { Authorization: `Bearer ${auth.value.access_token}` },
          signal: abortController.signal,
        })
        suggestions.value = (data.suggestions ?? []).slice(0, 3)
      }
      catch (e: unknown) {
        // Ignore aborted requests — a newer fetch is already in-flight
        if (e instanceof Error && e.name === 'AbortError') return
        suggestions.value = []
      }
      finally {
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
