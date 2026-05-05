import { ExpoConfig, ConfigContext } from 'expo/config';

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: 'Colloquia',
  slug: 'colloquia-mobile',
  scheme: 'colloquia',
  version: '0.1.0',
  orientation: 'portrait',
  userInterfaceStyle: 'dark',
  backgroundColor: '#121212',
  icon: './assets/icon.png',
  android: {
    package: 'com.colloquia.mobile',
    versionCode: 1,
    adaptiveIcon: {
      foregroundImage: './assets/adaptive-icon.png',
      backgroundColor: '#121212',
    },
    allowBackup: false,
  },
  plugins: [
    'expo-router',
    'expo-secure-store',
    'expo-local-authentication',
    [
      'expo-notifications',
      {
        color: '#ffb700',
        sounds: [],
      },
    ],
  ],
  extra: {
    router: {
      origin: false,
    },
    eas: {
      projectId: 'colloquia-poc',
    },
  },
});
