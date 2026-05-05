import React from 'react';
import { View, StyleSheet, ViewProps } from 'react-native';
import { colors, radius } from '@/theme/tokens';

interface CardProps extends ViewProps {
  children: React.ReactNode;
  padding?: number;
}

export function Card({ children, padding = 16, style, ...rest }: CardProps): React.ReactElement {
  return (
    <View
      style={[styles.card, { padding }, style]}
      {...rest}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.card,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
  },
});
