#!/usr/bin/env bash
# Instala o APK mais recente num dispositivo Android conectado via USB.
# Uso: ./scripts/install-apk.sh [debug|release]
set -euo pipefail

VARIANT="${1:-debug}"
MOBILE_DIR="$(dirname "$0")/.."
APK_PATH="$MOBILE_DIR/android/app/build/outputs/apk/$VARIANT/app-${VARIANT}.apk"

if [ ! -f "$APK_PATH" ]; then
  echo "❌  APK não encontrado: $APK_PATH"
  echo "    Constrói primeiro:"
  echo "      npm run build:apk:$VARIANT    (ou cd android && ./gradlew assemble$(echo "${VARIANT:0:1}" | tr '[:lower:]' '[:upper:]')${VARIANT:1})"
  exit 1
fi

if ! adb devices | grep -q "device$"; then
  echo "❌  Nenhum dispositivo Android detectado."
  echo "    Verifica: USB debugging activado, cabo ligado, 'adb devices' lista o device."
  exit 1
fi

echo "📲  A instalar $APK_PATH..."
adb install -r "$APK_PATH"
echo "✅  APK instalado com sucesso."
echo "    Se a app não abrir automaticamente: procura 'Colloquia' no launcher."
