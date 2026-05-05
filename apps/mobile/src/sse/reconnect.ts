type Listener<T> = (value: T) => void;

class SimpleEmitter<T> {
  private listeners: Set<Listener<T>> = new Set();

  emit(value: T): void {
    this.listeners.forEach((l) => l(value));
  }

  on(listener: Listener<T>): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }
}

// Emits the new access token on rotation, or null on logout.
export const tokenRotatedEmitter = new SimpleEmitter<string | null>();

export function calcBackoffMs(retry: number): number {
  return Math.min(2000 * Math.pow(2, retry), 30_000);
}
