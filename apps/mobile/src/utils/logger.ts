const REDACT_KEYS = ['authorization', 'refresh_token', 'access_token', 'token'];
const REDACT_URL_PARAMS = ['token'];

function redactObject(obj: unknown): unknown {
  if (typeof obj !== 'object' || obj === null) return obj;
  const result: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
    if (REDACT_KEYS.some((r) => k.toLowerCase().includes(r))) {
      result[k] = '[REDACTED]';
    } else {
      result[k] = redactObject(v);
    }
  }
  return result;
}

function redactUrl(url: string): string {
  try {
    const u = new URL(url);
    for (const param of REDACT_URL_PARAMS) {
      if (u.searchParams.has(param)) u.searchParams.set(param, '[REDACTED]');
    }
    return u.toString();
  } catch {
    return url;
  }
}

export const logger = {
  info: (msg: string, data?: unknown) => {
    if (__DEV__) console.info(`[INFO] ${msg}`, data ? redactObject(data) : '');
  },
  warn: (msg: string, data?: unknown) => {
    console.warn(`[WARN] ${msg}`, data ? redactObject(data) : '');
  },
  error: (msg: string, data?: unknown) => {
    console.error(`[ERROR] ${msg}`, data ? redactObject(data) : '');
  },
  sse: (msg: string, url: string) => {
    if (__DEV__) console.info(`[SSE] ${msg}`, redactUrl(url));
  },
};
