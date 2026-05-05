interface JwtPayload {
  sub: string;
  exp: number;
  iat: number;
  email?: string;
  jti?: string;
}

function base64UrlDecode(str: string): string {
  const padded = str.replace(/-/g, '+').replace(/_/g, '/');
  const rem = padded.length % 4;
  const paddedStr = rem === 0 ? padded : padded + '='.repeat(4 - rem);
  return atob(paddedStr);
}

export function decodeJwt(token: string): JwtPayload {
  const parts = token.split('.');
  if (parts.length !== 3) throw new Error('Invalid JWT format');
  const payload = JSON.parse(base64UrlDecode(parts[1]!)) as JwtPayload;
  return payload;
}

export function getAccessTokenExpMs(token: string): number {
  const { exp } = decodeJwt(token);
  return exp * 1000;
}
