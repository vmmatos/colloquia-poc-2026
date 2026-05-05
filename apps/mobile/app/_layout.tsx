import '../global.css';
import React, { useEffect, useState, useCallback } from 'react';
import { Stack } from 'expo-router';
import * as SplashScreen from 'expo-splash-screen';
import * as Notifications from 'expo-notifications';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { StatusBar } from 'expo-status-bar';
import {
  Inter_400Regular,
  Inter_500Medium,
  Inter_600SemiBold,
  Inter_700Bold,
  useFonts,
} from '@expo-google-fonts/inter';
import {
  SourceSerif4_400Regular_Italic,
} from '@expo-google-fonts/source-serif-4';
import { bootstrapAuth } from '@/auth/bootstrap';
import { logger } from '@/utils/logger';
import { BiometricLockScreen } from '@/components/system/BiometricLockScreen';
import { useAuthStore } from '@/auth/store';
import { useAppForegroundLock } from '@/hooks/useAppForegroundLock';
import { performRefresh } from '@/auth/bootstrap';
import { colors } from '@/theme/tokens';
import { ErrorBoundary } from '@/components/system/ErrorBoundary';

SplashScreen.preventAutoHideAsync();

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: false,
    shouldSetBadge: true,
  }),
});

function AppLock(): React.ReactElement | null {
  const [locked, setLocked] = useState(false);

  const handleLock = useCallback(() => {
    useAuthStore.getState().setUnauthenticated();
    setLocked(true);
  }, []);

  useAppForegroundLock({ onLock: handleLock });

  const handleUnlock = useCallback(async () => {
    try {
      await performRefresh();
      setLocked(false);
    } catch {
      // refresh failed; stay locked, auth store already set to unauthenticated
    }
  }, []);

  if (!locked) return null;
  return <BiometricLockScreen onUnlock={handleUnlock} />;
}

export default function RootLayout(): React.ReactElement | null {
  const [fontsLoaded, fontError] = useFonts({
    Inter_400Regular,
    Inter_500Medium,
    Inter_600SemiBold,
    Inter_700Bold,
    SourceSerif4_400Regular_Italic,
  });
  const [authReady, setAuthReady] = useState(false);

  useEffect(() => {
    async function init() {
      try {
        await bootstrapAuth();
      } catch (err) {
        logger.error('RootLayout: bootstrapAuth threw', err);
      } finally {
        setAuthReady(true);
        SplashScreen.hideAsync().catch((e) => logger.warn('RootLayout: hideAsync failed', e));
      }
    }
    init();
  }, []);

  // Proceed even if fonts fail — better a fallback font than a permanent black screen.
  if ((!fontsLoaded && !fontError) || !authReady) return null;

  return (
    <ErrorBoundary>
      <GestureHandlerRootView style={{ flex: 1, backgroundColor: colors.background }}>
        <SafeAreaProvider>
          <StatusBar style="light" backgroundColor={colors.background} />
          <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}>
            <Stack.Screen name="(auth)" />
            <Stack.Screen name="(app)" />
          </Stack>
          <AppLock />
        </SafeAreaProvider>
      </GestureHandlerRootView>
    </ErrorBoundary>
  );
}
