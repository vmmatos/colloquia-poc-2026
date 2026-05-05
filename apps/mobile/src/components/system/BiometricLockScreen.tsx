import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import * as LocalAuthentication from 'expo-local-authentication';
import * as ScreenCapture from 'expo-screen-capture';
import { colors, radius, spacing } from '@/theme/tokens';
import { pt } from '@/i18n/pt';

interface BiometricLockScreenProps {
  onUnlock: () => void;
}

export function BiometricLockScreen({ onUnlock }: BiometricLockScreenProps): React.ReactElement {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    ScreenCapture.preventScreenCaptureAsync();
    authenticate();
    return () => { ScreenCapture.allowScreenCaptureAsync(); };
  }, []);

  async function authenticate() {
    setLoading(true);
    setError(null);
    try {
      const result = await LocalAuthentication.authenticateAsync({
        promptMessage: pt.biometricPrompt,
        fallbackLabel: pt.biometricFallback,
        cancelLabel: pt.cancel,
      });
      if (result.success) {
        onUnlock();
      } else {
        setError('Autenticação cancelada.');
      }
    } catch {
      setError('Erro ao autenticar.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={styles.container}>
      <Text style={styles.appName}>Colloquia</Text>
      <Text style={styles.subtitle}>{pt.unlockApp}</Text>
      {loading ? (
        <ActivityIndicator color={colors.primary} size="large" style={styles.spinner} />
      ) : (
        <>
          {error ? <Text style={styles.error}>{error}</Text> : null}
          <TouchableOpacity style={styles.btn} onPress={authenticate}>
            <Text style={styles.btnText}>🔐  {pt.biometricPrompt}</Text>
          </TouchableOpacity>
        </>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: colors.background,
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.lg,
    padding: spacing.xl,
  },
  appName: {
    color: colors.primary,
    fontSize: 32,
    fontFamily: 'Inter_700Bold',
    letterSpacing: -0.5,
  },
  subtitle: {
    color: colors.mutedFg,
    fontSize: 16,
    fontFamily: 'Inter_400Regular',
  },
  spinner: { marginTop: spacing.xl },
  error: { color: colors.destructive, fontSize: 13 },
  btn: {
    backgroundColor: colors.secondary,
    borderRadius: radius.md,
    paddingHorizontal: 24,
    paddingVertical: 14,
    borderWidth: 1,
    borderColor: colors.border,
  },
  btnText: {
    color: colors.foreground,
    fontSize: 15,
    fontFamily: 'Inter_500Medium',
  },
});
