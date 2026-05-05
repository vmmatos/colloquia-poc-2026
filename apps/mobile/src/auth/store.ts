import { create } from 'zustand';

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated';

interface AuthState {
  status: AuthStatus;
  userId: string | null;
  accessToken: string | null;
  accessExpMs: number | null; // ms timestamp when access token expires
  setAuthenticated: (userId: string, accessToken: string, expMs: number) => void;
  setUnauthenticated: () => void;
  setLoading: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  status: 'loading',
  userId: null,
  accessToken: null,
  accessExpMs: null,

  setAuthenticated: (userId, accessToken, expMs) =>
    set({ status: 'authenticated', userId, accessToken, accessExpMs: expMs }),

  setUnauthenticated: () =>
    set({ status: 'unauthenticated', userId: null, accessToken: null, accessExpMs: null }),

  setLoading: () => set({ status: 'loading' }),
}));

export function getAuthStore() {
  return useAuthStore.getState();
}
