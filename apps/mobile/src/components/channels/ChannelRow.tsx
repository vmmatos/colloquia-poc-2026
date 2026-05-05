import React from 'react';
import { TouchableOpacity, View, Text, StyleSheet } from 'react-native';
import { colors, radius } from '@/theme/tokens';
import type { Channel } from '@/types/shared';
import { useChannelsStore } from '@/stores/channels';

interface ChannelRowProps {
  channel: Channel;
  active?: boolean;
  onPress: () => void;
  displayName?: string;
}

const typePrefix: Record<string, string> = {
  channel: '#',
  group: '⬡',
  dm: '@',
};

export function ChannelRow({ channel, active, onPress, displayName }: ChannelRowProps): React.ReactElement {
  const unread = useChannelsStore((s) => s.unread[channel.id] ?? 0);

  return (
    <TouchableOpacity
      onPress={onPress}
      activeOpacity={0.7}
      style={[styles.row, active && styles.activeRow]}
    >
      <Text style={styles.prefix} numberOfLines={1}>
        {typePrefix[channel.type] ?? '#'}
      </Text>
      <Text style={[styles.name, active && styles.activeName]} numberOfLines={1}>
        {displayName || channel.name || '…'}
      </Text>
      {unread > 0 && (
        <View style={styles.badge}>
          <Text style={styles.badgeText}>{unread > 9 ? '9+' : unread}</Text>
        </View>
      )}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    paddingHorizontal: 12,
    borderRadius: radius.md,
    gap: 6,
  },
  activeRow: {
    backgroundColor: colors.secondary,
  },
  prefix: {
    color: colors.mutedFg,
    fontSize: 14,
    fontFamily: 'Inter_500Medium',
    width: 16,
  },
  name: {
    flex: 1,
    color: colors.mutedFg,
    fontSize: 14,
    fontFamily: 'Inter_400Regular',
  },
  activeName: {
    color: colors.foreground,
    fontFamily: 'Inter_500Medium',
  },
  badge: {
    backgroundColor: colors.primary,
    borderRadius: radius.full,
    minWidth: 18,
    height: 18,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 4,
  },
  badgeText: {
    color: colors.primaryFg,
    fontSize: 10,
    fontFamily: 'Inter_700Bold',
  },
});
