import React, { useCallback, useRef } from 'react';
import { FlatList, ActivityIndicator, View, StyleSheet } from 'react-native';
import { MessageBubble, DisplayMessage } from './MessageBubble';
import { colors } from '@/theme/tokens';

interface MessageListProps {
  messages: DisplayMessage[];
  loadingMore: boolean;
  onEndReached: () => void;
  onRetryDraft: (tempId: string) => void;
}

export function MessageList({
  messages,
  loadingMore,
  onEndReached,
  onRetryDraft,
}: MessageListProps): React.ReactElement {
  const listRef = useRef<FlatList<DisplayMessage>>(null);

  const renderItem = useCallback(
    ({ item }: { item: DisplayMessage }) => (
      <MessageBubble
        msg={item}
        showAvatar
        onRetryDraft={onRetryDraft}
      />
    ),
    [onRetryDraft],
  );

  return (
    <FlatList
      ref={listRef}
      data={messages}
      renderItem={renderItem}
      keyExtractor={(item) =>
        item.type === 'draft' ? `draft-${item.tempId}` : (item as { id: string }).id
      }
      inverted
      onEndReached={onEndReached}
      onEndReachedThreshold={0.2}
      ListFooterComponent={
        loadingMore ? (
          <View style={styles.loader}>
            <ActivityIndicator color={colors.primary} size="small" />
          </View>
        ) : null
      }
      removeClippedSubviews={false}
    />
  );
}

const styles = StyleSheet.create({
  loader: { padding: 12, alignItems: 'center' },
});
