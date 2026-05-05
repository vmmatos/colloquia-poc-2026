#!/usr/bin/env bash
# Detecta o IP LAN do Mac e escreve o .env.local para o dispositivo físico.
# Uso: ./scripts/set-api-base.sh
set -euo pipefail

# Tenta en0 (Wi-Fi típico no Mac), depois en1.
IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo "")

if [ -z "$IP" ]; then
  echo "❌  Não foi possível detectar IP LAN. Verifica a ligação Wi-Fi."
  echo "    Alternativas:"
  echo "      Android emulator: http://10.0.2.2"
  echo "      iOS simulator:    http://localhost"
  exit 1
fi

ENV_FILE="$(dirname "$0")/../.env.local"
echo "EXPO_PUBLIC_API_BASE_URL=http://$IP" > "$ENV_FILE"
echo "✅  Escrito: EXPO_PUBLIC_API_BASE_URL=http://$IP -> .env.local"
echo "    Garante que o telemóvel e o Mac estão na mesma Wi-Fi (sem AP isolation)."
echo "    Stack local: cd dev && docker compose up"
