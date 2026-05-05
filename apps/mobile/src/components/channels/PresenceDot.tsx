import React from 'react';
import { View, StyleSheet } from 'react-native';
import { usePresenceStore } from '@/stores/presence';
import { colors } from '@/theme/tokens';

interface PresenceDotProps {
  userId: string;
  size?: number;
}

export function PresenceDot({ userId, size = 10 }: PresenceDotProps): React.ReactElement | null {
  const isOnline = usePresenceStore((s) => s.isOnline(userId));

  return (
    <View
      style={[
        styles.dot,
        {
          width: size,
          height: size,
          borderRadius: size / 2,
          backgroundColor: isOnline ? colors.online : colors.offline,
        },
      ]}
    />
  );
}

const styles = StyleSheet.create({
  dot: {
    borderWidth: 2,
    borderColor: colors.background,
  },
});
