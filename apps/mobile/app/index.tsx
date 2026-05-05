import React from 'react';
import { Redirect } from 'expo-router';
import { useAuthStore } from '@/auth/store';

export default function Index(): React.ReactElement {
  const status = useAuthStore((s) => s.status);

  if (status === 'loading') return <></>;
  if (status === 'authenticated') return <Redirect href="/(app)/channels" />;
  return <Redirect href="/(auth)/login" />;
}
