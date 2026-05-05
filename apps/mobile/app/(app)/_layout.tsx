import React, { useEffect, useRef } from 'react';
import { Redirect, Tabs } from 'expo-router';
import { useAuthStore } from '@/auth/store';
import { useSseLifecycle } from '@/hooks/useSseLifecycle';
import { useHeartbeat } from '@/hooks/useHeartbeat';
import { useChannelsStore } from '@/stores/channels';
import { useNotificationResponse, requestNotificationPermissions } from '@/hooks/useNotifications';
import { colors } from '@/theme/tokens';
import { Ionicons } from '@expo/vector-icons';

function AppServices() {
  useSseLifecycle();
  useHeartbeat();
  useNotificationResponse();

  const fetch = useChannelsStore((s) => s.fetch);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    fetch();
    void requestNotificationPermissions();
    timerRef.current = setInterval(() => fetch(), 30_000);
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [fetch]);

  return null;
}

export default function AppLayout(): React.ReactElement {
  const status = useAuthStore((s) => s.status);

  if (status === 'unauthenticated') return <Redirect href="/(auth)/login" />;

  return (
    <>
      <AppServices />
      <Tabs
        screenOptions={{
          headerShown: false,
          tabBarStyle: {
            backgroundColor: colors.sidebar,
            borderTopColor: colors.border,
            height: 64,
            paddingTop: 8,
            paddingBottom: 10,
          },
          tabBarActiveTintColor: colors.primary,
          tabBarInactiveTintColor: colors.mutedFg,
          tabBarLabelStyle: { fontSize: 11, fontFamily: 'Inter_500Medium' },
        }}
      >
        <Tabs.Screen
          name="channels"
          options={{
            title: 'Canais',
            tabBarIcon: ({ focused, color }) => (
              <Ionicons name={focused ? 'chatbubble' : 'chatbubble-outline'} size={22} color={color} />
            ),
          }}
        />
        <Tabs.Screen
          name="search"
          options={{
            title: 'Pesquisar',
            tabBarIcon: ({ focused, color }) => (
              <Ionicons name={focused ? 'search' : 'search-outline'} size={22} color={color} />
            ),
          }}
        />
        <Tabs.Screen
          name="profile"
          options={{
            title: 'Perfil',
            tabBarIcon: ({ focused, color }) => (
              <Ionicons name={focused ? 'person' : 'person-outline'} size={22} color={color} />
            ),
          }}
        />
      </Tabs>
    </>
  );
}
