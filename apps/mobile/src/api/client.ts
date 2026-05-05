import { getApiBaseUrl } from '@/utils/env';
import { getAuthStore } from '@/auth/store';
import { logger } from '@/utils/logger';

// Single in-flight refresh promise — all concurrent 401s await this.
let refreshPromise: Promise<void> | null = null;

// Set by bootstrap.ts after import to avoid circular dependency.
let performRefresh: (() => Promise<void>) | null = null;
let performLogout: (() => void) | null = null;

export function configureClient(opts: {
  refresh: () => Promise<void>;
  logout: () => void;
}): void {
  performRefresh = opts.refresh;
  performLogout = opts.logout;
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
    public readonly body?: unknown,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

interface RequestOptions extends Omit<RequestInit, 'headers'> {
  headers?: Record<string, string>;
  authRequired?: boolean;
  isRetry?: boolean;
}

export async function apiFetch<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { authRequired = true, isRetry = false, ...fetchOpts } = options;
  const base = getApiBaseUrl();
  const url = `${base}${path}`;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (authRequired) {
    const { accessToken } = getAuthStore();
    if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
  }

  logger.info(`apiFetch ${fetchOpts.method ?? 'GET'} ${path}`);

  const res = await fetch(url, { ...fetchOpts, headers });

  if (res.status === 401 && authRequired && !isRetry) {
    if (!refreshPromise && performRefresh) {
      refreshPromise = performRefresh().finally(() => {
        refreshPromise = null;
      });
    }

    try {
      await refreshPromise;
    } catch {
      performLogout?.();
      throw new ApiError(401, 'Session expired');
    }

    // Retry with new token.
    return apiFetch<T>(path, { ...options, isRetry: true });
  }

  if (!res.ok) {
    let body: unknown;
    try { body = await res.json(); } catch { /* ignore */ }
    const msg = (body as { message?: string })?.message ?? res.statusText;
    throw new ApiError(res.status, msg, body);
  }

  if (res.status === 204) return undefined as T;

  return res.json() as Promise<T>;
}
