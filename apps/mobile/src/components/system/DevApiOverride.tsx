import React, { useState } from 'react';
import { View, Text, Alert, StyleSheet } from 'react-native';
import { Sheet } from '@/components/primitives/Sheet';
import { Input } from '@/components/primitives/Input';
import { Button } from '@/components/primitives/Button';
import { getApiBaseUrl, setDevApiBaseUrl, clearDevApiBaseUrl } from '@/utils/env';
import { colors, spacing } from '@/theme/tokens';

interface DevApiOverrideProps {
  visible: boolean;
  onClose: () => void;
}

export function DevApiOverride({ visible, onClose }: DevApiOverrideProps): React.ReactElement {
  const [url, setUrl] = useState(getApiBaseUrl());

  function handleSave() {
    if (!url.trim()) return;
    setDevApiBaseUrl(url.trim());
    Alert.alert('Dev', `API base: ${url.trim()}\nAtivo imediatamente para novos pedidos.`);
    onClose();
  }

  function handleClear() {
    clearDevApiBaseUrl();
    setUrl(process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://10.0.2.2');
    Alert.alert('Dev', 'Override limpo. URL de ambiente restaurada.');
    onClose();
  }

  return (
    <Sheet visible={visible} onClose={onClose}>
      <Text style={styles.title}>API Base URL (dev)</Text>
      <Text style={styles.hint}>Actual: {getApiBaseUrl()}</Text>
      <View style={styles.form}>
        <Input
          value={url}
          onChangeText={setUrl}
          placeholder="http://192.168.1.XX"
          autoCapitalize="none"
          autoCorrect={false}
          keyboardType="url"
        />
        <Button label="Guardar" onPress={handleSave} />
        <Button label="Limpar override" onPress={handleClear} variant="ghost" />
      </View>
    </Sheet>
  );
}

const styles = StyleSheet.create({
  title: { color: colors.foreground, fontSize: 16, fontFamily: 'Inter_600SemiBold', marginBottom: 8 },
  hint: { color: colors.mutedFg, fontSize: 12, marginBottom: spacing.md },
  form: { gap: spacing.md, paddingBottom: 24 },
});
