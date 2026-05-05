import React, { useState } from 'react';
import {
  TextInput,
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  TextInputProps,
  ViewProps,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius } from '@/theme/tokens';

interface InputProps extends Omit<TextInputProps, 'style'> {
  label?: string;
  error?: string;
  secureToggle?: boolean;
  textStyle?: TextInputProps['style'];
  style?: ViewProps['style'];
}

export function Input({
  label,
  error,
  secureToggle,
  secureTextEntry,
  multiline,
  textStyle,
  style,
  ...rest
}: InputProps): React.ReactElement {
  const [hidden, setHidden] = useState(secureTextEntry ?? false);

  return (
    <View style={[styles.wrapper, style]}>
      {label ? <Text style={styles.label}>{label}</Text> : null}
      <View style={[styles.inputRow, multiline && styles.inputRowMultiline, error ? styles.inputError : styles.inputDefault]}>
        <TextInput
          style={[styles.input, textStyle]}
          placeholderTextColor={colors.mutedFg}
          selectionColor={colors.primary}
          cursorColor={colors.primary}
          secureTextEntry={hidden}
          autoCapitalize="none"
          autoCorrect={false}
          multiline={multiline}
          {...rest}
        />
        {secureToggle && (
          <TouchableOpacity onPress={() => setHidden((h) => !h)} style={styles.eye}>
            <Ionicons name={hidden ? 'eye-outline' : 'eye-off-outline'} size={18} color={colors.mutedFg} />
          </TouchableOpacity>
        )}
      </View>
      <Text style={styles.errorText}>{error ?? ''}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { gap: 4 },
  label: {
    color: colors.foreground,
    fontSize: 13,
    fontFamily: 'Inter_500Medium',
    marginBottom: 2,
  },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'center',
    borderRadius: radius.md,
    borderWidth: 1,
    backgroundColor: colors.secondary,
    paddingHorizontal: 12,
    height: 44,
  },
  inputRowMultiline: {
    height: undefined,
    minHeight: 44,
    alignItems: 'flex-start',
    paddingVertical: 10,
  },
  inputDefault: { borderColor: colors.border },
  inputError: { borderColor: colors.destructive },
  input: {
    flex: 1,
    color: colors.foreground,
    fontSize: 14,
    fontFamily: 'Inter_400Regular',
  },
  eye: { paddingLeft: 8 },
  errorText: {
    color: colors.destructive,
    fontSize: 11,
    fontFamily: 'Inter_400Regular',
    minHeight: 16,
  },
});
