import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  FlatList,
  ActivityIndicator,
  StyleSheet,
} from 'react-native';
import { Sheet } from '@/components/primitives/Sheet';
import { Input } from '@/components/primitives/Input';
import { Avatar } from '@/components/primitives/Avatar';
import { useDebouncedValue } from '@/hooks/useDebouncedValue';
import { usersApi } from '@/api/endpoints';
import { colors, radius, spacing } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import type { UserProfile } from '@/types/shared';

interface UserSearchModalProps {
  visible: boolean;
  onClose: () => void;
  onSelectUser: (user: UserProfile) => void;
  excludeUserId?: string;
  actionLabel?: string;
}

export function UserSearchModal({
  visible,
  onClose,
  onSelectUser,
  excludeUserId,
  actionLabel = pt.createDm,
}: UserSearchModalProps): React.ReactElement {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const debounced = useDebouncedValue(query, 300);

  useEffect(() => {
    if (!visible) { setQuery(''); setResults([]); }
  }, [visible]);

  useEffect(() => {
    if (debounced.length < 2) { setResults([]); return; }
    setLoading(true);
    usersApi.search(debounced)
      .then((res) => setResults(res.filter((u) => u.user_id !== excludeUserId)))
      .catch(() => setResults([]))
      .finally(() => setLoading(false));
  }, [debounced, excludeUserId]);

  return (
    <Sheet visible={visible} onClose={onClose} scrollable={false}>
      <Text style={styles.title}>{pt.searchUsersPlaceholder}</Text>
      <Input
        value={query}
        onChangeText={setQuery}
        placeholder={pt.searchUsersPlaceholder}
        autoFocus
      />
      <View style={styles.results}>
        {loading ? (
          <ActivityIndicator color={colors.primary} style={styles.loader} />
        ) : (
          <FlatList
            data={results}
            keyExtractor={(u) => u.user_id}
            keyboardShouldPersistTaps="handled"
            renderItem={({ item }) => (
              <TouchableOpacity
                style={styles.userRow}
                onPress={() => { onSelectUser(item); onClose(); }}
              >
                <Avatar name={item.name} size="md" />
                <View style={styles.userInfo}>
                  <Text style={styles.userName}>{item.name || item.email.split('@')[0]}</Text>
                  <Text style={styles.userEmail} numberOfLines={1}>{item.email}</Text>
                </View>
                <Text style={styles.action}>{actionLabel}</Text>
              </TouchableOpacity>
            )}
            ListEmptyComponent={
              debounced.length >= 2 ? (
                <Text style={styles.empty}>Sem resultados</Text>
              ) : null
            }
          />
        )}
      </View>
    </Sheet>
  );
}

const styles = StyleSheet.create({
  title: {
    color: colors.foreground,
    fontSize: 16,
    fontFamily: 'Inter_600SemiBold',
    marginBottom: spacing.md,
  },
  results: { marginTop: spacing.md, minHeight: 200 },
  loader: { marginTop: 24 },
  userRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 10,
    gap: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  userInfo: { flex: 1 },
  userName: { color: colors.foreground, fontSize: 14, fontFamily: 'Inter_500Medium' },
  userEmail: { color: colors.mutedFg, fontSize: 12, fontFamily: 'Inter_400Regular' },
  action: { color: colors.primary, fontSize: 13, fontFamily: 'Inter_500Medium' },
  empty: { color: colors.mutedFg, fontSize: 13, textAlign: 'center', marginTop: 24 },
});
