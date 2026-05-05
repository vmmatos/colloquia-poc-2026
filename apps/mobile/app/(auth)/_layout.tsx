import React from 'react';
import { Redirect, Stack } from 'expo-router';
import { useAuthStore } from '@/auth/store';
import { colors } from '@/theme/tokens';

export default function AuthLayout(): React.ReactElement {
  const status = useAuthStore((s) => s.status);

  if (status === 'authenticated') return <Redirect href="/(app)/channels" />;

  return (
    <Stack
      screenOptions={{
        headerShown: false,
        contentStyle: { backgroundColor: colors.background },
        animation: 'fade',
      }}
    />
  );
}
