import {
  Inter_400Regular,
  Inter_500Medium,
  Inter_600SemiBold,
  Inter_700Bold,
  useFonts as useInterFonts,
} from '@expo-google-fonts/inter';
import {
  SourceSerif4_400Regular_Italic,
  useFonts as useSerif4Fonts,
} from '@expo-google-fonts/source-serif-4';

export const fontAssets = {
  Inter_400Regular,
  Inter_500Medium,
  Inter_600SemiBold,
  Inter_700Bold,
  SourceSerif4_400Regular_Italic,
};

export const fontFamilies = {
  sans: 'Inter_400Regular',
  sansMedium: 'Inter_500Medium',
  sansSemibold: 'Inter_600SemiBold',
  sansBold: 'Inter_700Bold',
  serifItalic: 'SourceSerif4_400Regular_Italic',
} as const;

export { useInterFonts, useSerif4Fonts };
