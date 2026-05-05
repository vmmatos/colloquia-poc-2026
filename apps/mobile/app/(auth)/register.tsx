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
import { registerSchema } from '@/utils/validation';

export default function RegisterScreen(): React.ReactElement {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string; password?: string; general?: string }>({});

  React.useEffect(() => {
    ScreenCapture.preventScreenCaptureAsync();
    return () => { ScreenCapture.allowScreenCaptureAsync(); };
  }, []);

  async function handleRegister() {
    const parsed = registerSchema.safeParse({ email, password });
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
      const data = await authApi.register(email, password);
      await handleAuthSuccess(data);
    } catch (err) {
      const msg = err instanceof ApiError
        ? (err.status === 409 ? 'Email já registado.' : err.message)
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
        <Text style={styles.logo}>Colloquia</Text>
        <Text style={styles.tagline}>Cria a tua conta</Text>

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
            autoComplete="new-password"
            textContentType="newPassword"
            error={errors.password}
          />
          <Button
            label={loading ? pt.loading : pt.register}
            onPress={handleRegister}
            loading={loading}
            style={styles.btn}
          />
        </View>

        <Link href="/(auth)/login" asChild>
          <TouchableOpacity style={styles.switchLink}>
            <Text style={styles.switchText}>
              Já tens conta?{' '}
              <Text style={styles.switchAction}>Entrar</Text>
            </Text>
          </TouchableOpacity>
        </Link>
      </ScrollView>
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
  errorBannerText: { color: colors.destructive, fontSize: 13 },
  switchLink: { alignItems: 'center' },
  switchText: { color: colors.mutedFg, fontSize: 14 },
  switchAction: { color: colors.primary, fontFamily: 'Inter_500Medium' },
});
