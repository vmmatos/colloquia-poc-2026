import { useEffect } from 'react';
import { useAuthStore } from '@/auth/store';
import { useChannelsStore } from '@/stores/channels';
import { usePresenceStore } from '@/stores/presence';
import { messagesStream } from '@/sse/messagesStream';
import { presenceStream } from '@/sse/presenceStream';
import { useMessagesStore } from '@/stores/messages';
import { useUiStore } from '@/stores/ui';
import { messagesApi } from '@/api/endpoints';
import { logger } from '@/utils/logger';
import { sendLocalNotification } from '@/hooks/useNotifications';

export function useSseLifecycle(): void {
  const status = useAuthStore((s) => s.status);
  const accessToken = useAuthStore((s) => s.accessToken);
  const channels = useChannelsStore((s) => s.channels);
  const updatePresence = usePresenceStore((s) => s.update);
  const appendMessage = useMessagesStore((s) => s.appendMessage);
  const incrementUnread = useChannelsStore((s) => s.incrementUnread);

  // Open/close SSE when auth changes.
  useEffect(() => {
    if (status !== 'authenticated' || !accessToken) {
      messagesStream.close();
      presenceStream.close();
      return;
    }

    const channelIds = channels.map((c) => c.id);
    messagesStream.open(channelIds, accessToken);
    presenceStream.open(accessToken);

    const unsub = messagesStream.subscribe((msg) => {
      appendMessage(msg.channel_id, msg);
      if (msg.channel_id !== useUiStore.getState().activeChannelId) {
        incrementUnread(msg.channel_id);
        // Send local notification — advantage #2
        const ch = useChannelsStore.getState().channels.find((c) => c.id === msg.channel_id);
        if (ch) {
          void sendLocalNotification({
            title: `#${ch.name ?? ch.id.slice(0, 8)}`,
            body: (msg.content?.length ?? 0) > 100 ? (msg.content ?? '').slice(0, 97) + '…' : (msg.content ?? ''),
            channelId: msg.channel_id,
          });
        }
      }
    });

    const unsubPresence = presenceStream.subscribe(updatePresence);

    return () => {
      unsub();
      unsubPresence();
    };
  }, [status, accessToken]); // eslint-disable-line react-hooks/exhaustive-deps

  // Update subscribed channels when channel list changes.
  useEffect(() => {
    if (status !== 'authenticated') return;
    const channelIds = channels.map((c) => c.id);
    messagesStream.updateChannels(channelIds);
  }, [channels, status]);
}

// Refetches gap in messages + channels when returning from background.
export async function refetchAfterReconnect(activeChannelId: string | null): Promise<void> {
  const { fetch } = useChannelsStore.getState();
  await fetch();

  if (activeChannelId) {
    try {
      const msgs = await messagesApi.list(activeChannelId);
      useMessagesStore.getState().setHistory(activeChannelId, msgs);
    } catch (e) {
      logger.warn('refetchAfterReconnect: messages fetch failed', e);
    }
  }
}
