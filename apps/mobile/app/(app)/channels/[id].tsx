import React, { useEffect, useMemo, useCallback, useState } from 'react';
import { View, Text, StyleSheet, ActivityIndicator } from 'react-native';
import { useLocalSearchParams, Stack } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useMessagesStore } from '@/stores/messages';
import { useChannelsStore } from '@/stores/channels';
import { useAuthStore } from '@/auth/store';
import { useUiStore } from '@/stores/ui';
import { messagesApi, usersApi } from '@/api/endpoints';
import { MessageList } from '@/components/chat/MessageList';
import { Composer } from '@/components/chat/Composer';
import { EmptyState } from '@/components/primitives/EmptyState';
import { colors } from '@/theme/tokens';
import { pt } from '@/i18n/pt';
import type { DraftMessage } from '@/types/shared';
import type { DisplayMessage } from '@/components/chat/MessageBubble';
import { nowSecs } from '@/utils/time';
import { ApiError } from '@/api/client';
import { logger } from '@/utils/logger';
import { useAssist } from '@/hooks/useAssist';

const userNameCache: Record<string, string> = {};

export default function ChannelScreen(): React.ReactElement {
  const { id } = useLocalSearchParams<{ id: string }>();
  const insets = useSafeAreaInsets();
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  const setHistory = useMessagesStore((s) => s.setHistory);
  const prependMessages = useMessagesStore((s) => s.prependMessages);
  const addDraft = useMessagesStore((s) => s.addDraft);
  const updateDraftStatus = useMessagesStore((s) => s.updateDraftStatus);
  const replaceDraft = useMessagesStore((s) => s.replaceDraft);

  const channel = useChannelsStore((s) => s.channels.find((c) => c.id === id));
  const clearUnread = useChannelsStore((s) => s.clearUnread);
  const userId = useAuthStore((s) => s.userId);
  const setActiveChannel = useUiStore((s) => s.setActiveChannel);
  const { suggestions, isLoading: suggestionsLoading, debouncedFetch, clearSuggestions } = useAssist();

  const messages = useMessagesStore((s) => s.messages[id!]) ?? [];
  const drafts = useMessagesStore((s) => s.drafts[id!]) ?? [];
  const [nameMap, setNameMap] = useState<Record<string, string>>(userNameCache);

  useEffect(() => {
    setActiveChannel(id!);
    clearUnread(id!);
    return () => setActiveChannel(null);
  }, [id, setActiveChannel, clearUnread]);

  useEffect(() => {
    const unknownIds = [...new Set(messages.map((m) => m.user_id).filter(Boolean))]
      .filter((uid) => !userNameCache[uid]);
    if (unknownIds.length === 0) return;
    let cancelled = false;
    Promise.all(
      unknownIds.map((uid) =>
        usersApi.getById(uid)
          .then((p) => { userNameCache[uid] = p.name || (p.email ?? '').split('@')[0]; })
          .catch(() => { userNameCache[uid] = (uid ?? '').slice(0, 8); })
      )
    ).then(() => { if (!cancelled) setNameMap({ ...userNameCache }); });
    return () => { cancelled = true; };
  }, [messages]);

  useEffect(() => {
    if (!id) return;
    messagesApi.list(id)
      .then((msgs) => { setHistory(id, msgs); setLoadError(null); })
      .catch((err) => {
        logger.warn('ChannelScreen: initial messages fetch failed', err);
        setLoadError('Erro a carregar mensagens');
      })
      .finally(() => setLoading(false));
  }, [id, setHistory]);

  const handleEndReached = useCallback(async () => {
    if (loadingMore || !hasMore || messages.length === 0) return;
    const oldest = messages[0];
    if (!oldest) return;
    setLoadingMore(true);
    try {
      const older = await messagesApi.list(id!, { beforeId: oldest.id });
      if (older.length < 50) setHasMore(false);
      prependMessages(id!, older);
    } catch (err) {
      logger.warn('ChannelScreen: load-more fetch failed', err);
    } finally {
      setLoadingMore(false);
    }
  }, [loadingMore, hasMore, messages, id, prependMessages]);

  function handleComposerTextChange(text: string) {
    debouncedFetch(id!, text);
  }

  async function sendMessage(content: string) {
    clearSuggestions();
    const tempId = `draft-${Date.now()}-${Math.random()}`;
    const draft: DraftMessage = {
      tempId,
      channel_id: id!,
      content,
      status: 'pending',
      created_at: nowSecs(),
    };
    addDraft(draft);
    try {
      const msg = await messagesApi.send(id!, content);
      replaceDraft(id!, tempId, msg);
    } catch (err) {
      if (err instanceof ApiError) {
        // Got a response from server (e.g., 400, 500) — failed, not recoverable
        updateDraftStatus(tempId, 'failed');
      }
      // Network error (TypeError) — keep pending for manual retry when online
    }
  }

  async function retryDraft(tempId: string) {
    const draft = drafts.find((d) => d.tempId === tempId);
    if (!draft) return;
    updateDraftStatus(tempId, 'pending');
    try {
      const msg = await messagesApi.send(id!, draft.content);
      replaceDraft(id!, tempId, msg);
    } catch {
      updateDraftStatus(tempId, 'failed');
    }
  }

  // Build display messages: real messages + drafts merged by time
  const displayMessages = useMemo((): DisplayMessage[] => {
    try {
      const real: DisplayMessage[] = messages.map((m) => ({
        ...m,
        type: 'sent',
        isOwn: m.user_id === userId,
        authorName: nameMap[m.user_id] ?? (m.user_id ?? '').slice(0, 8),
      }));
      const draftDisplay: DisplayMessage[] = drafts.map((d) => ({
        ...d,
        type: 'draft',
        authorName: 'Eu',
      }));
      // inverted list: newest at index 0
      return [...draftDisplay.reverse(), ...real.slice().reverse()];
    } catch (e) {
      logger.error('displayMessages computation error', e);
      return [];
    }
  }, [messages, drafts, userId, nameMap]);

  // Send local notifications for incoming messages in other channels
  useEffect(() => {
    // Notification handler for messages arriving in this channel (we're already here, so do nothing)
    // Channel-level suppression handled in useSseLifecycle via activeChannelId
  }, []);

  return (
    <View style={[styles.container, { paddingBottom: insets.bottom }]}>
      <Stack.Screen
        options={{
          title: channel
            ? (channel.type === 'dm' ? '@ ' : '# ') + (channel.name || '…')
            : '…',
          headerBackTitle: 'Canais',
        }}
      />
      {loading ? (
        <View style={styles.center}>
          <ActivityIndicator color={colors.primary} />
        </View>
      ) : loadError ? (
        <EmptyState title={loadError} subtitle="Verifica a ligação e tenta novamente" />
      ) : displayMessages.length === 0 ? (
        <EmptyState title={pt.noMessages} />
      ) : (
        <MessageList
          messages={displayMessages}
          loadingMore={loadingMore}
          onEndReached={handleEndReached}
          onRetryDraft={retryDraft}
        />
      )}
      <Composer
        onSend={sendMessage}
        onTextChange={handleComposerTextChange}
        onSuggestionApplied={clearSuggestions}
        suggestions={suggestions}
        suggestionsLoading={suggestionsLoading}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
});
