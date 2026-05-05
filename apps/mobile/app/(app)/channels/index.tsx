import React, { useState, useCallback, useEffect, useRef } from 'react';
import {
  View,
  Text,
  FlatList,
  RefreshControl,
  TouchableOpacity,
  StyleSheet,
} from 'react-native';
import { useRouter, Stack } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useChannelsStore } from '@/stores/channels';
import { useUiStore } from '@/stores/ui';
import { useAuthStore } from '@/auth/store';
import { channelsApi, usersApi } from '@/api/endpoints';
import { ChannelRow } from '@/components/channels/ChannelRow';
import { CreateChannelModal } from '@/components/modals/CreateChannelModal';
import { EmptyState } from '@/components/primitives/EmptyState';
import { colors, spacing } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import type { Channel } from '@/types/shared';

export default function ChannelsScreen(): React.ReactElement {
  const router = useRouter();
  const insets = useSafeAreaInsets();
  const channels = useChannelsStore((s) => s.channels);
  const loading = useChannelsStore((s) => s.loading);
  const error = useChannelsStore((s) => s.error);
  const fetch = useChannelsStore((s) => s.fetch);
  const clearUnread = useChannelsStore((s) => s.clearUnread);
  const activeChannelId = useUiStore((s) => s.activeChannelId);
  const [showCreate, setShowCreate] = useState(false);
  const [dmNames, setDmNames] = useState<Record<string, string>>({});
  const resolvedRef = useRef<Set<string>>(new Set());
  const myUserId = useAuthStore((s) => s.userId);

  const regularChannels = channels.filter((c) => c.type === 'channel');
  const dms = channels.filter((c) => c.type === 'dm' || c.type === 'group');

  useEffect(() => {
    const unresolved = dms.filter((ch) => !ch.name && !resolvedRef.current.has(ch.id));
    if (unresolved.length === 0) return;
    unresolved.forEach((ch) => resolvedRef.current.add(ch.id));
    Promise.all(
      unresolved.map(async (ch): Promise<[string, string]> => {
        try {
          const members = await channelsApi.getMembers(ch.id);
          const other = members.find((m) => m.user_id !== myUserId);
          if (!other) return [ch.id, 'DM'];
          const profile = await usersApi.getById(other.user_id);
          return [ch.id, profile.name || profile.email.split('@')[0]];
        } catch {
          return [ch.id, 'DM'];
        }
      })
    ).then((entries) => setDmNames((prev) => ({ ...prev, ...Object.fromEntries(entries) })));
  }, [dms, myUserId]);

  const handleChannelPress = useCallback(
    (ch: Channel) => {
      clearUnread(ch.id);
      router.push(`/channels/${ch.id}`);
    },
    [clearUnread, router],
  );

  const handleCreated = useCallback(
    (channelId: string) => {
      router.push(`/channels/${channelId}`);
    },
    [router],
  );

  return (
    <View style={styles.container}>
      <Stack.Screen
        options={{
          title: 'Canais',
          headerRight: () => (
            <TouchableOpacity onPress={() => setShowCreate(true)} style={styles.addBtn}>
              <Text style={styles.addBtnText}>＋</Text>
            </TouchableOpacity>
          ),
        }}
      />
      {error && channels.length === 0 && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>{error}</Text>
          <TouchableOpacity onPress={fetch}>
            <Text style={styles.errorRetry}>Tentar novamente</Text>
          </TouchableOpacity>
        </View>
      )}
      <FlatList
        style={styles.list}
        data={[]}
        renderItem={() => null}
        ListHeaderComponent={
          channels.length === 0 && !loading && !error ? (
            <EmptyState
              title={pt.noChannels}
              subtitle="Cria um canal para começar"
              action={{ label: pt.newChannel, onPress: () => setShowCreate(true) }}
            />
          ) : (
            <View>
              {regularChannels.length > 0 && (
                <View>
                  <Text style={styles.sectionHeader}>{pt.channels.toUpperCase()}</Text>
                  {regularChannels.map((ch) => (
                    <ChannelRow
                      key={ch.id}
                      channel={ch}
                      active={ch.id === activeChannelId}
                      onPress={() => handleChannelPress(ch)}
                    />
                  ))}
                </View>
              )}
              {dms.length > 0 && (
                <View style={styles.section}>
                  <Text style={styles.sectionHeader}>{pt.directMessages.toUpperCase()}</Text>
                  {dms.map((ch) => (
                    <ChannelRow
                      key={ch.id}
                      channel={ch}
                      displayName={ch.name || dmNames[ch.id]}
                      active={ch.id === activeChannelId}
                      onPress={() => handleChannelPress(ch)}
                    />
                  ))}
                </View>
              )}
            </View>
          )
        }
        keyExtractor={() => 'header'}
        refreshControl={
          <RefreshControl
            refreshing={loading}
            onRefresh={fetch}
            tintColor={colors.primary}
            colors={[colors.primary]}
          />
        }
        contentContainerStyle={styles.listContent}
      />
      <CreateChannelModal
        visible={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={handleCreated}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  list: { flex: 1 },
  listContent: { padding: spacing.md, flexGrow: 1 },
  sectionHeader: {
    color: colors.mutedFg,
    fontSize: 11,
    fontFamily: 'Inter_600SemiBold',
    letterSpacing: 0.8,
    paddingVertical: 8,
    paddingHorizontal: 12,
  },
  section: { marginTop: spacing.md },
  addBtn: { marginRight: 8, padding: 8 },
  addBtnText: { color: colors.primary, fontSize: 22, fontFamily: 'Inter_400Regular' },
  errorBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: `${colors.destructive}22`,
    borderBottomWidth: 1,
    borderBottomColor: colors.destructive,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.sm,
  },
  errorText: { flex: 1, color: colors.destructive, fontSize: 13, fontFamily: 'Inter_400Regular' },
  errorRetry: { color: colors.primary, fontSize: 13, fontFamily: 'Inter_500Medium' },
});
