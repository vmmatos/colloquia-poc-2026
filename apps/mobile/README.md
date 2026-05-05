# Colloquia Mobile

App Android nativa (Expo SDK 52 + React Native 0.76) para o Colloquia POC.

## 3 vantagens vs. web

| # | Vantagem | Detalhe |
|---|---|---|
| 1 | **Biometric lock** | Refresh token em SecureStore (Keychain/EncryptedSharedPreferences). App bloqueia após 5 min em background e requer FaceID/fingerprint para desbloquear. |
| 2 | **Live notifications + deep-link** | Mensagens novas em canais não-activos geram notificações OS nativas enquanto a app está activa. Tap abre o canal directo via `colloquia://channel/<id>`. |
| 3 | **Offline cache + queued sends** | Últimas 50 msgs por canal persistidas em MMKV. Mensagens compostas offline ficam em fila e são enviadas automáticamente quando a ligação retorna. |

---

## Pré-requisitos

- Node 20+
- npm
- Android Studio ou Android SDK (para `adb`)
- Java 17 (para Gradle)
- Stack local a correr: `cd dev && docker compose up`

---

## Setup

```bash
cd apps/mobile
npm install
```

### Configurar URL do backend

#### Dispositivo físico (mesma Wi-Fi que o Mac):
```bash
./scripts/set-api-base.sh   # detecta IP LAN e escreve .env.local
```
Ou manualmente:
```bash
echo "EXPO_PUBLIC_API_BASE_URL=http://192.168.1.XX" > .env.local
```

#### Android emulator:
```bash
echo "EXPO_PUBLIC_API_BASE_URL=http://10.0.2.2" > .env.local
```

> **Nota:** Usa **porta 80** (ingress NGINX), não a porta 8000 (KrakenD directo).  
> O `nginx.conf` do dev faz bypass do KrakenD para `/api/v1/messages/stream` (SSE).

---

## Desenvolvimento

```bash
npm run prebuild        # gera android/ (uma vez)
npm run android         # compila, instala no device/emulator e inicia Metro
```

### Override da API em runtime (debug)

No ecrã de Perfil, toca 5× na versão da app → modal "API Base URL (dev)".

---

## Build APK local

```bash
# Debug APK (instala com adb)
npm run build:apk:debug

# Instalar no device ligado
./scripts/install-apk.sh debug

# Ou via adb directamente
adb install -r android/app/build/outputs/apk/debug/app-debug.apk
```

---

## Ícone da App

Design em `assets/icon.svg` — dois balões de fala em ouro `#ffb700` dispostos na diagonal sobre fundo escuro `#121212`, representando uma conversa entre dois participantes.

Para gerar os ficheiros PNG necessários pelo Expo:

```bash
npm install --save-dev sharp   # uma vez
npm run generate:icons          # gera assets/icon.png e assets/adaptive-icon.png
npm run prebuild                # re-aplica ao android/ (apaga android/ antes se já existir)
```

Faz commit dos PNGs gerados em `assets/`. O SVG original serve como fonte de verdade do design.

---

## GitHub Actions

O workflow `.github/workflows/mobile-android-apk.yml` corre em:
- `push` neste branch quando `apps/mobile/**` muda
- `workflow_dispatch` (manual, com input `api_base_url`)
- `release` publicado — anexa o APK ao release para sideload

O APK é **debug-signed** (sem keystore secrets). Funciona para sideload;  
Play Protect irá avisar "app não verificada" — aceitável para POC.

Para upgrade para release-signed: adicionar 4 secrets ao repo GitHub:
- `ANDROID_KEYSTORE_BASE64`
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

E a variável `MOBILE_API_BASE_URL` (URL pública do backend).

---

## Instalar APK de um GitHub Release

1. Va ao tab "Releases" do repo no GitHub
2. Faz download do `app-debug.apk`
3. No Android: Definições → Segurança → "Fontes desconhecidas" → Permite
4. Abre o ficheiro APK para instalar

---

## Estrutura

```
app/              expo-router routes
  _layout.tsx     Bootstrap, fontes, lock screen
  (auth)/         Login, Registo
  (app)/          App protegida (tabs)
    channels/     Lista + chat
    search.tsx    Pesquisa de utilizadores
    profile.tsx   Perfil + logout
src/
  api/            HTTP client + endpoints
  auth/           JWT decode, SecureStore, refresh scheduler
  sse/            SSE streams (mensagens + presence)
  stores/         Zustand stores
  components/     Primitivos + chat + modais + sistema
  hooks/          useHeartbeat, useSseLifecycle, useNotifications, ...
  theme/          Tokens de cor/tipografia (espelho da web)
  utils/          env, logger, time, validation
  i18n/           Strings em português
```

---

## Segurança

- Refresh token em `expo-secure-store` (`WHEN_UNLOCKED_THIS_DEVICE_ONLY`)
- Access token apenas em memória (Zustand), nunca em disco
- Biometric lock após 5 min em background
- `expo-screen-capture` em ecrãs sensíveis (login, lock)
- Logger redacta `Authorization`, `refresh_token`, `?token=` URLs
- Cleartext HTTP apenas em builds de debug (`network_security_config.xml`)
- Validação client-side com zod espelha limites do BE

---

## Verificação E2E

Ver secção "Verification plan" no ficheiro de plano do projecto.

Comandos úteis:
```bash
adb logcat -s ReactNativeJS:*          # logs da app
adb shell run-as com.colloquia.mobile ls files/secure-store   # debug SecureStore
adb reverse tcp:18080 tcp:80           # redirecionar porta NGINX (alternativa ao LAN IP)
```
