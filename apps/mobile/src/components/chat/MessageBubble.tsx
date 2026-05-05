import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { Avatar } from '@/components/primitives/Avatar';
import { colors, radius, spacing } from '@/theme/tokens';
import { formatTime } from '@/utils/time';
import type { Message, DraftMessage } from '@/types/shared';
import { pt } from '@/i18n/pt';

export type DisplayMessage =
  | (Message & { type: 'sent'; isOwn: boolean; authorName: string; authorAvatar?: string })
  | (DraftMessage & { type: 'draft'; authorName: string });

interface MessageBubbleProps {
  msg: DisplayMessage;
  showAvatar: boolean;
  onRetryDraft?: (tempId: string) => void;
}

export function MessageBubble({ msg, showAvatar, onRetryDraft }: MessageBubbleProps): React.ReactElement {
  const isDraft = msg.type === 'draft';
  const isPending = isDraft && msg.status === 'pending';
  const isFailed = isDraft && msg.status === 'failed';

  const content = isDraft ? msg.content : (msg as Message).content;
  const time = isDraft
    ? formatTime(msg.created_at)
    : formatTime((msg as Message).created_at);

  return (
    <View style={styles.row}>
      <View style={styles.avatarSlot}>
        {showAvatar && (
          <Avatar
            name={msg.authorName}
            uri={(msg as { authorAvatar?: string }).authorAvatar}
            size="sm"
          />
        )}
      </View>
      <View style={styles.bubble}>
        {showAvatar && (
          <View style={styles.metaRow}>
            <Text style={styles.authorName}>{msg.authorName}</Text>
            <Text style={styles.time}>{time}</Text>
          </View>
        )}
        <Text style={[styles.content, isFailed && styles.failedContent]}>{content}</Text>
        {isPending && <Text style={styles.status}>{pt.messagePending}</Text>}
        {isFailed && (
          <TouchableOpacity
            onPress={() => isDraft && onRetryDraft?.((msg as DraftMessage).tempId)}
          >
            <Text style={styles.retryText}>{pt.messageFailed}</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: spacing.sm,
  },
  avatarSlot: { width: 24, marginTop: 2 },
  bubble: { flex: 1, gap: 2 },
  metaRow: { flexDirection: 'row', alignItems: 'baseline', gap: 8 },
  authorName: {
    color: colors.foreground,
    fontSize: 13,
    fontFamily: 'Inter_600SemiBold',
  },
  time: {
    color: colors.mutedFg,
    fontSize: 11,
    fontFamily: 'Inter_400Regular',
  },
  content: {
    color: colors.foreground,
    fontSize: 14,
    fontFamily: 'Inter_400Regular',
    lineHeight: 20,
  },
  failedContent: { opacity: 0.5 },
  status: {
    color: colors.mutedFg,
    fontSize: 11,
    fontFamily: 'Inter_400Regular',
    fontStyle: 'italic',
  },
  retryText: {
    color: colors.destructive,
    fontSize: 11,
    fontFamily: 'Inter_500Medium',
  },
});
