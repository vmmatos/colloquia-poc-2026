import { useEffect, useRef } from 'react';
import * as Notifications from 'expo-notifications';
import { useRouter } from 'expo-router';
import { logger } from '@/utils/logger';

export async function requestNotificationPermissions(): Promise<boolean> {
  const { status: existing } = await Notifications.getPermissionsAsync();
  if (existing === 'granted') return true;
  const { status } = await Notifications.requestPermissionsAsync();
  return status === 'granted';
}

export async function sendLocalNotification(opts: {
  title: string;
  body: string;
  channelId: string;
}): Promise<void> {
  try {
    await Notifications.scheduleNotificationAsync({
      content: {
        title: opts.title,
        body: opts.body,
        data: { channelId: opts.channelId },
        sound: false,
      },
      trigger: null, // immediate
    });
  } catch (err) {
    logger.warn('sendLocalNotification failed', err);
  }
}

export function useNotificationResponse(): void {
  const router = useRouter();
  const lastNotificationId = useRef<string | null>(null);

  useEffect(() => {
    const sub = Notifications.addNotificationResponseReceivedListener((response) => {
      const id = response.notification.request.identifier;
      if (id === lastNotificationId.current) return;
      lastNotificationId.current = id;

      const data = response.notification.request.content.data;
      const channelId = (data as { channelId?: string })?.channelId;
      if (channelId) {
        router.push(`/channels/${channelId}`);
      }
    });
    return () => sub.remove();
  }, [router]);
}
