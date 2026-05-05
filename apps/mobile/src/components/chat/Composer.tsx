import React, { useState, useRef } from 'react';
import {
  View,
  TextInput,
  TouchableOpacity,
  Text,
  StyleSheet,
  Platform,
  ScrollView,
  ActivityIndicator,
} from 'react-native';
import * as Haptics from 'expo-haptics';
import { colors, radius, spacing } from '@/theme/tokens';
import { pt } from '@/i18n/pt';

interface ComposerProps {
  onSend: (content: string) => void;
  onTextChange?: (text: string) => void;
  onSuggestionApplied?: () => void;
  suggestions?: string[];
  suggestionsLoading?: boolean;
  disabled?: boolean;
}

export function Composer({
  onSend,
  onTextChange,
  onSuggestionApplied,
  suggestions = [],
  suggestionsLoading = false,
  disabled,
}: ComposerProps): React.ReactElement {
  const [text, setText] = useState('');
  const inputRef = useRef<TextInput>(null);

  function handleTextChange(val: string) {
    setText(val);
    onTextChange?.(val);
  }

  function handleSend() {
    const trimmed = text.trim();
    if (!trimmed || disabled) return;
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
    onSend(trimmed);
    setText('');
    onTextChange?.('');
  }

  function applySuggestion(suggestion: string) {
    setText(suggestion);
    onSuggestionApplied?.();
    inputRef.current?.focus();
  }

  const canSend = text.trim().length > 0 && !disabled;
  const showSuggestions = suggestions.length > 0 || suggestionsLoading;

  return (
    <View>
      {showSuggestions && (
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          style={styles.pillsRow}
          contentContainerStyle={styles.pillsContent}
          keyboardShouldPersistTaps="always"
        >
          {suggestionsLoading && suggestions.length === 0 && (
            <ActivityIndicator size="small" color={colors.primary} style={styles.pillLoader} />
          )}
          {suggestions.map((s, i) => (
            <TouchableOpacity
              key={i}
              style={styles.pill}
              onPress={() => applySuggestion(s)}
              activeOpacity={0.7}
            >
              <Text style={styles.pillText} numberOfLines={1}>{s}</Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      )}
      <View style={styles.container}>
        <TextInput
          ref={inputRef}
          value={text}
          onChangeText={handleTextChange}
          placeholder={pt.typeMessage}
          placeholderTextColor={colors.mutedFg}
          selectionColor={colors.primary}
          style={styles.input}
          multiline
          maxLength={4000}
          onSubmitEditing={Platform.OS === 'ios' ? undefined : handleSend}
          blurOnSubmit={false}
        />
        <TouchableOpacity
          onPress={handleSend}
          disabled={!canSend}
          style={[styles.sendBtn, canSend ? styles.sendBtnActive : styles.sendBtnDisabled]}
          hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
        >
          <Text style={styles.sendIcon}>↑</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  pillsRow: {
    backgroundColor: colors.card,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    maxHeight: 68,
  },
  pillsContent: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    gap: 8,
    flexDirection: 'row',
    alignItems: 'center',
  },
  pill: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: radius.full,
    borderWidth: 1.5,
    borderColor: colors.primary,
    backgroundColor: '#2e2a1a',
    maxWidth: 260,
  },
  pillText: {
    color: '#ffffff',
    fontSize: 14,
    fontFamily: 'Inter_600SemiBold',
  },
  pillLoader: {
    marginRight: 4,
  },
  container: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    gap: 8,
    paddingHorizontal: 12,
    paddingVertical: 8,
    backgroundColor: colors.card,
    borderTopWidth: 1,
    borderTopColor: colors.border,
  },
  input: {
    flex: 1,
    minHeight: 40,
    maxHeight: 120,
    backgroundColor: colors.secondary,
    borderRadius: radius.md,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: 12,
    paddingVertical: 10,
    color: colors.foreground,
    fontSize: 14,
    fontFamily: 'Inter_400Regular',
  },
  sendBtn: {
    width: 40,
    height: 40,
    borderRadius: radius.full,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 2,
  },
  sendBtnActive: { backgroundColor: colors.primary },
  sendBtnDisabled: { backgroundColor: colors.secondary },
  sendIcon: {
    fontSize: 18,
    color: colors.primaryFg,
    fontFamily: 'Inter_700Bold',
  },
});
