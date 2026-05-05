import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { colors, spacing, radius } from '@/theme/tokens';
import { logger } from '@/utils/logger';

interface Props { children: React.ReactNode; }
interface State { hasError: boolean; message: string; stack: string; }

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, message: '', stack: '' };
  }

  static getDerivedStateFromError(error: unknown): Partial<State> {
    return { hasError: true, message: error instanceof Error ? error.message : String(error) };
  }

  componentDidCatch(error: unknown, info: React.ErrorInfo): void {
    this.setState({ stack: info.componentStack ?? '' });
    logger.error('ErrorBoundary caught an error', { error, componentStack: info.componentStack });
  }

  handleReset = (): void => { this.setState({ hasError: false, message: '', stack: '' }); };

  render(): React.ReactNode {
    if (!this.state.hasError) return this.props.children;
    return (
      <View style={styles.container}>
        <Text style={styles.title}>Algo correu mal</Text>
        <Text style={styles.subtitle}>A aplicação encontrou um erro inesperado.</Text>
        {__DEV__ && (
          <View style={styles.errorBox}>
            <Text style={styles.errorText} selectable>{this.state.message}</Text>
            {!!this.state.stack && (
              <Text style={[styles.errorText, styles.stackText]} selectable numberOfLines={15}>
                {this.state.stack.trim()}
              </Text>
            )}
          </View>
        )}
        <TouchableOpacity style={styles.btn} onPress={this.handleReset}>
          <Text style={styles.btnText}>Tentar novamente</Text>
        </TouchableOpacity>
      </View>
    );
  }
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background, alignItems: 'center', justifyContent: 'center', padding: spacing.xl, gap: spacing.lg },
  title: { color: colors.foreground, fontSize: 22, fontFamily: 'Inter_700Bold', textAlign: 'center' },
  subtitle: { color: colors.mutedFg, fontSize: 14, fontFamily: 'Inter_400Regular', textAlign: 'center' },
  errorBox: { backgroundColor: `${colors.destructive}22`, borderWidth: 1, borderColor: colors.destructive, borderRadius: radius.md, padding: spacing.md, width: '100%', gap: 8 },
  errorText: { color: colors.destructive, fontSize: 12, fontFamily: 'Inter_400Regular' },
  stackText: { fontSize: 10, opacity: 0.7 },
  btn: { backgroundColor: colors.secondary, borderRadius: radius.md, paddingHorizontal: 24, paddingVertical: 14, borderWidth: 1, borderColor: colors.border },
  btnText: { color: colors.foreground, fontSize: 15, fontFamily: 'Inter_500Medium' },
});
