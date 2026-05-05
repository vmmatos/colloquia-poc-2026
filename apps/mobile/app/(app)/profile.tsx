import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  Switch,
  StyleSheet,
  Alert,
} from 'react-native';
import { Stack } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import * as LocalAuthentication from 'expo-local-authentication';
import Constants from 'expo-constants';
import { usersApi } from '@/api/endpoints';
import { handleLogout } from '@/auth/bootstrap';
import { Input } from '@/components/primitives/Input';
import { Button } from '@/components/primitives/Button';
import { Avatar } from '@/components/primitives/Avatar';
import { DevApiOverride } from '@/components/system/DevApiOverride';
import { useUiStore } from '@/stores/ui';
import { useSettingsStore } from '@/stores/settings';
import { colors, spacing, radius } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import type { UserProfile } from '@/types/shared';
import { ApiError } from '@/api/client';
import { logger } from '@/utils/logger';

export default function ProfileScreen(): React.ReactElement {
  const insets = useSafeAreaInsets();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [name, setName] = useState('');
  const [bio, setBio] = useState('');
  const [saving, setSaving] = useState(false);
  const [logoutLoading, setLogoutLoading] = useState(false);
  const [showDevOverride, setShowDevOverride] = useState(false);
  const [biometricAvailable, setBiometricAvailable] = useState(false);
  const incrementDevTap = useUiStore((s) => s.incrementDevTap);
  const devTapCount = useUiStore((s) => s.devApiTapCount);
  const resetDevTap = useUiStore((s) => s.resetDevTap);
  const biometricLockEnabled = useSettingsStore((s) => s.biometricLockEnabled);
  const setBiometricLock = useSettingsStore((s) => s.setBiometricLock);

  useEffect(() => {
    usersApi.getProfile().then((p) => {
      setProfile(p);
      setName(p.name ?? '');
      setBio(p.bio ?? '');
    }).catch((err) => { logger.warn('ProfileScreen: getProfile failed', err); });

    LocalAuthentication.hasHardwareAsync().then(async (hasHw) => {
      if (!hasHw) return;
      const enrolled = await LocalAuthentication.isEnrolledAsync();
      setBiometricAvailable(enrolled);
    });
  }, []);

  async function handleSave() {
    setSaving(true);
    try {
      const updated = await usersApi.updateProfile({ name, bio });
      setProfile(updated);
      Alert.alert('Perfil', 'Guardado com sucesso.');
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : 'Erro ao guardar';
      Alert.alert('Erro', msg);
    } finally {
      setSaving(false);
    }
  }

  async function handleLogoutPress() {
    Alert.alert('Terminar sessão', 'Tens a certeza?', [
      { text: 'Cancelar', style: 'cancel' },
      {
        text: 'Terminar sessão',
        style: 'destructive',
        onPress: async () => {
          setLogoutLoading(true);
          await handleLogout();
          setLogoutLoading(false);
        },
      },
    ]);
  }

  async function handleBiometricToggle(value: boolean) {
    if (value) {
      // Require a successful biometric check before enabling.
      const result = await LocalAuthentication.authenticateAsync({
        promptMessage: 'Confirmar para ativar bloqueio biométrico',
        cancelLabel: 'Cancelar',
      });
      if (!result.success) return;
    }
    setBiometricLock(value);
  }

  function handleVersionTap() {
    incrementDevTap();
    if (devTapCount >= 4) {
      resetDevTap();
      setShowDevOverride(true);
    }
  }

  const version = Constants.expoConfig?.version ?? '0.1.0';

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <Stack.Screen options={{ headerShown: false }} />
      <ScrollView contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">
        <Text style={styles.title}>{pt.profile}</Text>

        <View style={styles.avatarRow}>
          <Avatar name={profile?.name ?? '?'} size="xl" />
          <View>
            <Text style={styles.email}>{profile?.email ?? ''}</Text>
            <Text style={styles.userId} selectable numberOfLines={1}>
              {profile?.user_id ?? ''}
            </Text>
          </View>
        </View>

        <View style={styles.form}>
          <Input
            label={pt.name}
            value={name}
            onChangeText={setName}
            maxLength={100}
          />
          <Input
            label={pt.bio}
            value={bio}
            onChangeText={setBio}
            maxLength={500}
            multiline
            textStyle={{ minHeight: 80, textAlignVertical: 'top', paddingTop: 10 }}
          />
          <Button label={saving ? pt.loading : pt.save} onPress={handleSave} loading={saving} />
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Segurança</Text>
          <View style={styles.settingRow}>
            <View style={styles.settingInfo}>
              <Text style={styles.settingLabel}>Bloqueio biométrico</Text>
              <Text style={styles.settingDesc}>
                {biometricAvailable
                  ? 'Bloqueia a app após 5 min em background'
                  : 'Biometria não disponível neste dispositivo'}
              </Text>
            </View>
            <Switch
              value={biometricLockEnabled}
              onValueChange={handleBiometricToggle}
              disabled={!biometricAvailable}
              trackColor={{ false: colors.border, true: colors.primary }}
              thumbColor={colors.foreground}
            />
          </View>
        </View>

        <Button
          label={logoutLoading ? pt.loading : pt.logout}
          onPress={handleLogoutPress}
          variant="danger"
          loading={logoutLoading}
          style={styles.logoutBtn}
        />

        <TouchableOpacity onPress={handleVersionTap} style={styles.versionRow}>
          <Text style={styles.version}>{pt.appVersion} {version}</Text>
        </TouchableOpacity>
      </ScrollView>

      <DevApiOverride visible={showDevOverride} onClose={() => setShowDevOverride(false)} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: spacing.xl, gap: spacing.lg },
  title: { color: colors.foreground, fontSize: 24, fontFamily: 'Inter_700Bold' },
  avatarRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  email: { color: colors.foreground, fontSize: 14, fontFamily: 'Inter_500Medium' },
  userId: { color: colors.mutedFg, fontSize: 11, fontFamily: 'Inter_400Regular', maxWidth: 200 },
  form: { gap: spacing.md },
  section: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.md,
    overflow: 'hidden',
  },
  sectionTitle: {
    color: colors.mutedFg,
    fontSize: 11,
    fontFamily: 'Inter_600SemiBold',
    letterSpacing: 0.8,
    textTransform: 'uppercase',
    paddingHorizontal: spacing.md,
    paddingTop: spacing.md,
    paddingBottom: spacing.sm,
  },
  settingRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    backgroundColor: colors.card,
  },
  settingInfo: { flex: 1, marginRight: spacing.md },
  settingLabel: { color: colors.foreground, fontSize: 14, fontFamily: 'Inter_500Medium' },
  settingDesc: { color: colors.mutedFg, fontSize: 12, fontFamily: 'Inter_400Regular', marginTop: 2 },
  logoutBtn: { marginTop: spacing.md },
  versionRow: { alignItems: 'center', marginTop: spacing.md },
  version: { color: colors.mutedFg, fontSize: 12 },
});
