import { create } from 'zustand';

interface UiState {
  activeChannelId: string | null;
  setActiveChannel: (id: string | null) => void;
  devApiTapCount: number;
  incrementDevTap: () => void;
  resetDevTap: () => void;
}

export const useUiStore = create<UiState>((set) => ({
  activeChannelId: null,
  setActiveChannel: (id) => set({ activeChannelId: id }),
  devApiTapCount: 0,
  incrementDevTap: () => set((s) => ({ devApiTapCount: s.devApiTapCount + 1 })),
  resetDevTap: () => set({ devApiTapCount: 0 }),
}));
