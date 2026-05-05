import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { Stack, useRouter } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { Input } from '@/components/primitives/Input';
import { Avatar } from '@/components/primitives/Avatar';
import { useDebouncedValue } from '@/hooks/useDebouncedValue';
import { usersApi, channelsApi } from '@/api/endpoints';
import { useChannelsStore } from '@/stores/channels';
import { colors, spacing } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import { useAuthStore } from '@/auth/store';
import type { UserProfile } from '@/types/shared';

export default function SearchScreen(): React.ReactElement {
  const insets = useSafeAreaInsets();
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const debounced = useDebouncedValue(query, 300);
  const myUserId = useAuthStore((s) => s.userId);
  const upsertChannel = useChannelsStore((s) => s.upsertChannel);

  useEffect(() => {
    if (debounced.length < 2) { setResults([]); return; }
    setLoading(true);
    usersApi.search(debounced, 30)
      .then((res) => setResults(res.filter((u) => u.user_id !== myUserId)))
      .catch(() => setResults([]))
      .finally(() => setLoading(false));
  }, [debounced, myUserId]);

  async function openDm(user: UserProfile) {
    try {
      const ch = await channelsApi.createDm(user.user_id);
      upsertChannel(ch);
      router.push(`/channels/${ch.id}`);
    } catch {
      // ignore
    }
  }

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <Stack.Screen options={{ headerShown: false }} />
      <View style={styles.header}>
        <Text style={styles.title}>{pt.search}</Text>
      </View>
      <View style={styles.inputWrap}>
        <Input
          value={query}
          onChangeText={setQuery}
          placeholder={pt.searchUsersPlaceholder}
          autoCorrect={false}
        />
      </View>
      {loading && <ActivityIndicator color={colors.primary} style={styles.loader} />}
      <FlatList
        data={results}
        keyExtractor={(u) => u.user_id}
        keyboardShouldPersistTaps="handled"
        renderItem={({ item }) => (
          <TouchableOpacity style={styles.userRow} onPress={() => openDm(item)}>
            <Avatar name={item.name} size="md" />
            <View style={styles.userInfo}>
              <Text style={styles.userName}>{item.name || item.email.split('@')[0]}</Text>
              <Text style={styles.userEmail} numberOfLines={1}>{item.email}</Text>
            </View>
            <Text style={styles.dmAction}>{pt.createDm}</Text>
          </TouchableOpacity>
        )}
        ListEmptyComponent={
          debounced.length >= 2 && !loading ? (
            <Text style={styles.empty}>Sem resultados</Text>
          ) : null
        }
        contentContainerStyle={styles.list}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  header: { paddingHorizontal: spacing.md, paddingTop: spacing.md },
  title: { color: colors.foreground, fontSize: 24, fontFamily: 'Inter_700Bold' },
  inputWrap: { padding: spacing.md },
  loader: { marginTop: 24 },
  list: { paddingHorizontal: spacing.md },
  userRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    gap: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  userInfo: { flex: 1 },
  userName: { color: colors.foreground, fontSize: 14, fontFamily: 'Inter_500Medium' },
  userEmail: { color: colors.mutedFg, fontSize: 12 },
  dmAction: { color: colors.primary, fontSize: 13, fontFamily: 'Inter_500Medium' },
  empty: { color: colors.mutedFg, textAlign: 'center', marginTop: 32, fontSize: 14 },
});
