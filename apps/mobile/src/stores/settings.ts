import { create } from 'zustand';
import { MMKV } from 'react-native-mmkv';

const storage = new MMKV({ id: 'settings' });

interface SettingsState {
  biometricLockEnabled: boolean;
  setBiometricLock: (enabled: boolean) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  biometricLockEnabled: storage.getBoolean('biometricLockEnabled') ?? false,

  setBiometricLock: (enabled) => {
    storage.set('biometricLockEnabled', enabled);
    set({ biometricLockEnabled: enabled });
  },
}));
