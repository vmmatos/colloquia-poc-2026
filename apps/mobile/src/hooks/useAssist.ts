import { useState, useRef, useCallback } from 'react';
import { assistApi } from '@/api/endpoints';

const DEBOUNCE_MS = 500;
const MIN_LENGTH = 10;
const TIMEOUT_MS = 10_000;
const MAX_SUGGESTIONS = 3;

export function useAssist() {
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const clearSuggestions = useCallback(() => {
    if (debounceRef.current !== null) {
      clearTimeout(debounceRef.current);
      debounceRef.current = null;
    }
    abortRef.current?.abort();
    abortRef.current = null;
    setSuggestions([]);
    setIsLoading(false);
  }, []);

  const debouncedFetch = useCallback((channelId: string, inputText: string) => {
    if (debounceRef.current !== null) {
      clearTimeout(debounceRef.current);
      debounceRef.current = null;
    }

    if (inputText.length <= MIN_LENGTH) {
      setSuggestions([]);
      setIsLoading(false);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      debounceRef.current = null;

      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      const timeoutId = setTimeout(() => controller.abort(), TIMEOUT_MS);
      setIsLoading(true);

      try {
        const data = await assistApi.suggestions(channelId, inputText);
        if (!controller.signal.aborted) {
          setSuggestions((data.suggestions ?? []).slice(0, MAX_SUGGESTIONS));
        }
      } catch {
        if (!controller.signal.aborted) {
          setSuggestions([]);
        }
      } finally {
        clearTimeout(timeoutId);
        if (!controller.signal.aborted) {
          setIsLoading(false);
          abortRef.current = null;
        }
      }
    }, DEBOUNCE_MS);
  }, []);

  return { suggestions, isLoading, debouncedFetch, clearSuggestions };
}
