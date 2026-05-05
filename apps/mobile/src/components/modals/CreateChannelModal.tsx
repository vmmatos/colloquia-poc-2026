import React, { useState } from 'react';
import {
  View,
  Text,
  Switch,
  TouchableOpacity,
  FlatList,
  StyleSheet,
  Alert,
  ScrollView,
} from 'react-native';
import { Sheet } from '@/components/primitives/Sheet';
import { Input } from '@/components/primitives/Input';
import { Button } from '@/components/primitives/Button';
import { Avatar } from '@/components/primitives/Avatar';
import { useDebouncedValue } from '@/hooks/useDebouncedValue';
import { usersApi, channelsApi } from '@/api/endpoints';
import { useChannelsStore } from '@/stores/channels';
import { colors, spacing, radius } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import type { UserProfile } from '@/types/shared';
import { ApiError } from '@/api/client';

interface CreateChannelModalProps {
  visible: boolean;
  onClose: () => void;
  onCreated: (channelId: string) => void;
}

export function CreateChannelModal({
  visible,
  onClose,
  onCreated,
}: CreateChannelModalProps): React.ReactElement {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<UserProfile[]>([]);
  const [selectedMembers, setSelectedMembers] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debounced = useDebouncedValue(searchQuery, 300);
  const upsertChannel = useChannelsStore((s) => s.upsertChannel);

  const [searching, setSearching] = useState(false);
  React.useEffect(() => {
    if (debounced.length < 2) { setSearchResults([]); return; }
    setSearching(true);
    usersApi.search(debounced)
      .then(setSearchResults)
      .catch(() => setSearchResults([]))
      .finally(() => setSearching(false));
  }, [debounced]);

  function reset() {
    setName(''); setDescription(''); setIsPrivate(false);
    setSearchQuery(''); setSearchResults([]); setSelectedMembers([]);
    setError(null);
  }

  async function handleCreate() {
    if (!name.trim()) { setError('Nome obrigatório'); return; }
    setLoading(true); setError(null);
    try {
      const ch = await channelsApi.create({
        name: name.trim(),
        description: description.trim() || undefined,
        is_private: isPrivate,
        type: 'channel',
        initial_member_ids: selectedMembers.map((m) => m.user_id),
      });
      upsertChannel(ch);
      reset();
      onCreated(ch.id);
      onClose();
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : 'Erro ao criar canal';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  function toggleMember(user: UserProfile) {
    setSelectedMembers((prev) =>
      prev.some((m) => m.user_id === user.user_id)
        ? prev.filter((m) => m.user_id !== user.user_id)
        : [...prev, user],
    );
  }

  return (
    <Sheet visible={visible} onClose={onClose} scrollable>
      <Text style={styles.title}>{pt.newChannel}</Text>
      <View style={styles.form}>
        <Input
          label={pt.channelName}
          value={name}
          onChangeText={setName}
          maxLength={80}
          error={error ?? undefined}
        />
        <Input
          label={pt.channelDescription}
          value={description}
          onChangeText={setDescription}
          maxLength={500}
        />
        <View style={styles.toggleRow}>
          <Text style={styles.label}>{pt.privateChannel}</Text>
          <Switch
            value={isPrivate}
            onValueChange={setIsPrivate}
            trackColor={{ true: colors.primary, false: colors.border }}
            thumbColor={colors.foreground}
          />
        </View>
        <Text style={styles.label}>{pt.addMembers}</Text>
        <Input
          value={searchQuery}
          onChangeText={setSearchQuery}
          placeholder={pt.searchUsersPlaceholder}
        />
        {searchResults.length > 0 && (
          <FlatList
            data={searchResults}
            scrollEnabled={false}
            keyExtractor={(u) => u.user_id}
            renderItem={({ item }) => {
              const selected = selectedMembers.some((m) => m.user_id === item.user_id);
              return (
                <TouchableOpacity style={styles.userRow} onPress={() => toggleMember(item)}>
                  <Avatar name={item.name} size="sm" />
                  <Text style={styles.userName}>{item.name}</Text>
                  {selected && <Text style={styles.checkmark}>✓</Text>}
                </TouchableOpacity>
              );
            }}
          />
        )}
        {selectedMembers.length > 0 && (
          <View style={styles.chips}>
            {selectedMembers.map((m) => (
              <TouchableOpacity
                key={m.user_id}
                style={styles.chip}
                onPress={() => toggleMember(m)}
              >
                <Text style={styles.chipText}>{m.name} ×</Text>
              </TouchableOpacity>
            ))}
          </View>
        )}
        <Button label={pt.create} onPress={handleCreate} loading={loading} />
      </View>
    </Sheet>
  );
}

const styles = StyleSheet.create({
  title: {
    color: colors.foreground,
    fontSize: 17,
    fontFamily: 'Inter_600SemiBold',
    marginBottom: spacing.md,
  },
  form: { gap: spacing.md, paddingBottom: 32 },
  label: { color: colors.foreground, fontSize: 13, fontFamily: 'Inter_500Medium' },
  toggleRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  userRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    gap: 10,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  userName: { flex: 1, color: colors.foreground, fontSize: 14 },
  checkmark: { color: colors.primary, fontSize: 16 },
  chips: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  chip: {
    backgroundColor: colors.secondary,
    borderRadius: radius.full,
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderWidth: 1,
    borderColor: colors.border,
  },
  chipText: { color: colors.foreground, fontSize: 12, fontFamily: 'Inter_500Medium' },
});
