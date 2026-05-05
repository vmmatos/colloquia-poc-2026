import React, { useState } from 'react';
import {
  View,
  Text,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
} from 'react-native';
import * as ScreenCapture from 'expo-screen-capture';
import { Link } from 'expo-router';
import { authApi } from '@/api/endpoints';
import { handleAuthSuccess } from '@/auth/bootstrap';
import { Input } from '@/components/primitives/Input';
import { Button } from '@/components/primitives/Button';
import { colors, spacing } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import { ApiError } from '@/api/client';
import { loginSchema } from '@/utils/validation';
import { DevApiOverride } from '@/components/system/DevApiOverride';
import { useUiStore } from '@/stores/ui';

export default function LoginScreen(): React.ReactElement {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string; password?: string; general?: string }>({});
  const [showDevOverride, setShowDevOverride] = useState(false);
  const incrementDevTap = useUiStore((s) => s.incrementDevTap);
  const devTapCount = useUiStore((s) => s.devApiTapCount);
  const resetDevTap = useUiStore((s) => s.resetDevTap);

  function handleLogoTap() {
    incrementDevTap();
    if (devTapCount >= 4) {
      resetDevTap();
      setShowDevOverride(true);
    }
  }

  React.useEffect(() => {
    ScreenCapture.preventScreenCaptureAsync();
    return () => { ScreenCapture.allowScreenCaptureAsync(); };
  }, []);

  async function handleLogin() {
    const parsed = loginSchema.safeParse({ email, password });
    if (!parsed.success) {
      const fieldErrors = parsed.error.flatten().fieldErrors;
      setErrors({
        email: fieldErrors.email?.[0],
        password: fieldErrors.password?.[0],
      });
      return;
    }
    setErrors({});
    setLoading(true);
    try {
      const data = await authApi.login(email, password);
      await handleAuthSuccess(data);
    } catch (err) {
      const msg = err instanceof ApiError
        ? (err.status === 401 ? 'Email ou palavra-passe incorrectos.' : err.message)
        : 'Erro de rede. Verifica a ligação.';
      setErrors({ general: msg });
    } finally {
      setLoading(false);
    }
  }

  return (
    <KeyboardAvoidingView
      style={styles.flex}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
        <TouchableOpacity onPress={handleLogoTap} activeOpacity={1}>
          <Text style={styles.logo}>Colloquia</Text>
        </TouchableOpacity>
        <Text style={styles.tagline}>A tua equipa, em qualquer lugar</Text>

        <View style={styles.form}>
          {errors.general ? (
            <View style={styles.errorBanner}>
              <Text style={styles.errorBannerText}>{errors.general}</Text>
            </View>
          ) : null}

          <Input
            label={pt.email}
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
            autoComplete="email"
            textContentType="emailAddress"
            error={errors.email}
          />
          <Input
            label={pt.password}
            value={password}
            onChangeText={setPassword}
            secureTextEntry
            secureToggle
            autoComplete="current-password"
            textContentType="password"
            error={errors.password}
          />
          <Button
            label={loading ? pt.loading : pt.login}
            onPress={handleLogin}
            loading={loading}
            style={styles.btn}
          />
        </View>

        <Link href="/(auth)/register" asChild>
          <TouchableOpacity style={styles.switchLink}>
            <Text style={styles.switchText}>
              Não tens conta?{' '}
              <Text style={styles.switchAction}>Registar</Text>
            </Text>
          </TouchableOpacity>
        </Link>
      </ScrollView>
      <DevApiOverride visible={showDevOverride} onClose={() => setShowDevOverride(false)} />
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.background },
  container: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: spacing.xl,
    gap: spacing.xl,
  },
  logo: {
    color: colors.primary,
    fontSize: 36,
    fontFamily: 'Inter_700Bold',
    letterSpacing: -1,
    textAlign: 'center',
  },
  tagline: {
    color: colors.mutedFg,
    fontSize: 14,
    fontFamily: 'SourceSerif4_400Regular_Italic',
    textAlign: 'center',
    marginTop: -spacing.md,
  },
  form: { gap: spacing.md },
  btn: { marginTop: spacing.sm },
  errorBanner: {
    backgroundColor: `${colors.destructive}22`,
    borderWidth: 1,
    borderColor: colors.destructive,
    borderRadius: 6,
    padding: 12,
  },
  errorBannerText: {
    color: colors.destructive,
    fontSize: 13,
    fontFamily: 'Inter_400Regular',
  },
  switchLink: { alignItems: 'center' },
  switchText: { color: colors.mutedFg, fontSize: 14 },
  switchAction: { color: colors.primary, fontFamily: 'Inter_500Medium' },
});
