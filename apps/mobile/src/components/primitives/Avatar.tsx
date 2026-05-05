import React from 'react';
import { View, Text, Image } from 'react-native';
import { colors } from '@/theme/tokens';

type AvatarSize = 'sm' | 'md' | 'lg' | 'xl';

const sizeMap: Record<AvatarSize, number> = {
  sm: 24,
  md: 32,
  lg: 40,
  xl: 80,
};

const fontSizeMap: Record<AvatarSize, number> = {
  sm: 9,
  md: 12,
  lg: 16,
  xl: 28,
};

interface AvatarProps {
  uri?: string;
  name?: string;
  size?: AvatarSize;
  online?: boolean | null;
}

export function Avatar({ uri, name, size = 'md', online }: AvatarProps): React.ReactElement {
  const dim = sizeMap[size];
  const fontSize = fontSizeMap[size];
  const initial = name ? name.charAt(0).toUpperCase() : '?';
  const dotSize = size === 'sm' ? 8 : 12;

  return (
    <View style={{ width: dim, height: dim }}>
      {uri ? (
        <Image
          source={{ uri }}
          style={{
            width: dim,
            height: dim,
            borderRadius: dim / 2,
            backgroundColor: colors.secondary,
          }}
        />
      ) : (
        <View
          style={{
            width: dim,
            height: dim,
            borderRadius: dim / 2,
            backgroundColor: colors.secondary,
            borderWidth: 1,
            borderColor: colors.border,
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Text style={{ color: colors.foreground, fontSize, fontFamily: 'Inter_500Medium' }}>
            {initial}
          </Text>
        </View>
      )}
      {online !== undefined && online !== null && (
        <View
          style={{
            position: 'absolute',
            bottom: 0,
            right: 0,
            width: dotSize,
            height: dotSize,
            borderRadius: dotSize / 2,
            backgroundColor: online ? colors.online : colors.offline,
            borderWidth: 2,
            borderColor: colors.background,
          }}
        />
      )}
    </View>
  );
}
