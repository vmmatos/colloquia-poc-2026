import React from 'react';
import { Stack } from 'expo-router';
import { colors } from '@/theme/tokens';

export default function ChannelsLayout(): React.ReactElement {
  return (
    <Stack
      screenOptions={{
        headerStyle: { backgroundColor: colors.sidebar },
        headerTintColor: colors.foreground,
        headerTitleStyle: { fontFamily: 'Inter_600SemiBold', fontSize: 16 },
        headerShadowVisible: false,
        contentStyle: { backgroundColor: colors.background },
      }}
    />
  );
}
