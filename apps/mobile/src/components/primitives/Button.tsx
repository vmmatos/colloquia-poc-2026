import React from 'react';
import { TouchableOpacity, Text, ActivityIndicator, StyleSheet, TouchableOpacityProps } from 'react-native';
import { colors, radius } from '@/theme/tokens';
import { pt } from '@/i18n/pt';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';

const variantStyles: Record<Variant, { bg: string; text: string; border?: string }> = {
  primary: { bg: colors.primary, text: colors.primaryFg },
  secondary: { bg: colors.secondary, text: colors.foreground },
  ghost: { bg: 'transparent', text: colors.foreground },
  danger: { bg: colors.destructive, text: '#ffffff' },
};

interface ButtonProps extends Omit<TouchableOpacityProps, 'style'> {
  label?: string;
  variant?: Variant;
  loading?: boolean;
  size?: 'sm' | 'md';
  style?: TouchableOpacityProps['style'];
}

export function Button({
  label,
  variant = 'primary',
  loading = false,
  size = 'md',
  disabled,
  children,
  style,
  ...rest
}: ButtonProps): React.ReactElement {
  const { bg, text } = variantStyles[variant];
  const isDisabled = disabled || loading;

  return (
    <TouchableOpacity
      activeOpacity={0.75}
      disabled={isDisabled}
      style={[
        styles.base,
        size === 'sm' && styles.sm,
        { backgroundColor: bg },
        isDisabled && styles.disabled,
        style,
      ]}
      {...rest}
    >
      {loading ? (
        <ActivityIndicator color={text} size="small" />
      ) : children ? (
        children
      ) : (
        <Text style={[styles.label, { color: text }, size === 'sm' && styles.labelSm]}>
          {loading ? pt.loading : label}
        </Text>
      )}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  base: {
    height: 44,
    paddingHorizontal: 20,
    borderRadius: radius.md,
    alignItems: 'center',
    justifyContent: 'center',
    flexDirection: 'row',
    gap: 8,
  },
  sm: {
    height: 34,
    paddingHorizontal: 14,
  },
  disabled: {
    opacity: 0.5,
  },
  label: {
    fontFamily: 'Inter_600SemiBold',
    fontSize: 14,
  },
  labelSm: {
    fontSize: 13,
  },
});
