# Frontend — Colloquia Web SPA

**Nuxt 4 single-page application with real-time messaging, presence tracking, and AI-powered reply suggestions.**

A dark-themed, responsive team messaging interface built with Vue 3, Tailwind CSS, and Nuxt 4. Supports English and Portuguese. Handles authentication via httpOnly cookie, real-time updates via SSE, and background token refresh.

---

## Tech Stack

| Concern | Technology | Version |
|---------|-----------|---------|
| **Framework** | Nuxt | ^4.3.1 |
| **UI Library** | Vue | ^3.5.29 |
| **Router** | vue-router | ^4.6.4 |
| **Styling** | Tailwind CSS | ^6.14.0 |
| **i18n** | @nuxtjs/i18n | ^10.2.4 |
| **Emoji Picker** | emoji-picker-element | ^1.29.1 |
| **Testing** | Vitest + @nuxt/test-utils | ^3.2.0 |
| **Package Manager** | yarn (+ package-lock.json) | — |

---

## Project Structure

```
app/
├── app.vue                    # Root component
├── assets/css/main.css        # Global styles
├── components/
│   ├── CreateChannelModal.vue    # New channel creation
│   ├── ManageChannelModal.vue    # Members / settings
│   ├── NewDMModal.vue            # Start DM search
│   ├── NewGroupModal.vue         # Upgrade DM to group
│   ├── NotificationCenter.vue    # Bell + toast queue
│   ├── ProfilePanel.vue          # Edit own profile
│   ├── MessageToast.vue          # Background message toasts
│   └── ui/                       # UI primitives
│       ├── Avatar.vue
│       ├── Button.vue
│       ├── Card.vue
│       └── Input.vue
├── composables/
│   ├── useAuth.ts                # Login/register/logout state
│   ├── useAuthFetch.ts           # Auto-retry on 401 + refresh
│   ├── useAssist.ts              # AI reply suggestions (debounced)
│   ├── useChannels.ts            # Channel list + CRUD
│   ├── useDMPeers.ts             # Map DM channel → other user
│   ├── useIsMobile.ts            # Responsive media query
│   ├── useLocale.ts              # i18n + persist to localStorage
│   ├── useMessageStore.ts        # In-memory message cache
│   ├── useMessaging.ts           # Fetch/send messages
│   ├── useNotifications.ts       # Toast + mention state
│   ├── usePresence.ts            # SSE presence stream
│   ├── useSSE.ts                 # SSE message stream
│   ├── useTokenRefresh.ts        # Auto-refresh before expiry
│   └── useUsersCache.ts          # Lazy user profile cache
├── layouts/
│   ├── auth.vue                  # Centered login/register
│   └── default.vue               # Full app shell + sidebar
├── middleware/
│   ├── auth.ts                   # Redirect to /login if unauthenticated
│   └── guest.ts                  # Redirect to / if already authenticated
├── pages/
│   ├── index.vue                 # Empty state + redirect to first channel
│   ├── login.vue                 # Login form
│   ├── register.vue              # Registration form
│   ├── profile.vue               # Trigger profile panel
│   ├── channels/[id].vue         # Main messaging view
│   └── debug.vue                 # Dev/testing page
└── plugins/
    ├── auth.ts                   # Token refresh on app load
    └── i18n-hydrate.client.ts    # Locale resolution
```

**Top-level files:**
```
nuxt.config.ts           # Nuxt configuration
package.json
tailwind.config.ts       # Dark theme + custom animations
tsconfig.json
vitest.config.ts         # Test configuration

server/api/              # BFF proxy routes (httpOnly cookies, redirects)
shared/types/            # TypeScript types (auth, channels, etc.)
i18n/locales/            # en.json, pt.json (English + Portuguese)
```

---

## How the App Works

### Authentication Flow

1. **App Load** → `plugins/auth.ts` runs
   - If no access token in state, try `refreshToken()` using httpOnly cookie
   - Set `authReady = true` (unblocks middleware)
   
2. **User navigates** → middleware runs
   - `auth` middleware: redirect to `/login` if not authenticated
   - `guest` middleware: redirect to `/` if already authenticated
   
3. **User logs in** → `useAuth().login()`
   - POST `/api/auth/login` (BFF proxy)
   - BFF gets response, sets httpOnly `refresh_token` cookie, returns access token
   - Frontend stores access token in state

4. **Authenticated requests** → `useAuthFetch()` adds `Authorization: Bearer <token>` header
   - On 401: automatically call `refreshToken()`, retry once
   - On second 401: clear state, redirect to `/login`

5. **Token expiry** → `useTokenRefresh` watches expiry
   - 60 seconds before expiry: schedule a refresh
   - On page close: timer is cleaned up

### Layout & Navigation

**`auth` layout (login/register pages):**
- Centered single column
- Heading + tagline
- Login/register form

**`default` layout (main app):**
- Fixed left sidebar (60% width on desktop, collapsible on mobile)
- Channels section (public channels, DMs)
- Unread count badge per channel
- "New Channel" + "New DM" buttons
- Main content area (right side)
- Overlay modals: CreateChannelModal, ManageChannelModal, NewDMModal, NewGroupModal, ProfilePanel
- Message toast stack (bottom-right)

### Real-Time Updates

1. **SSE Connections:**
   - `useSSE()` opens `/api/messages/stream?channel_id=...&channel_id=...&token=...`
   - Subscribes to all user's channels at once (single connection)
   - On channel list change, reconnects with new channels (exponential backoff)
   
   - `usePresence()` opens `/api/users/presence/stream?token=...`
   - Receives online/offline events
   - Sends heartbeat every 10 seconds
   
2. **Message Handling:**
   - `useMessageStore` is the single source of truth for messages
   - On history fetch: `setHistory()` merges fetched messages with live SSE events (race-safe)
   - On SSE event: `append()` adds to store (deduplicates by ID)
   - UI watches `messageStore` and re-renders

3. **Presence Tracking:**
   - `usePresence` maintains a `presenceMap: Record<userId, boolean>`
   - Components (Avatar) render a green dot if `presenceMap[userId]`
   - Also used for user status in header

### AI Suggestions

`useAssist` (in channels/[id].vue):
- Debounces input on every keystroke (500ms delay)
- Skips inputs ≤10 characters
- Sends POST to `/api/assist/suggestions` with context
- Request timeout: 10 seconds + AbortController
- Displays up to 3 suggestions as interactive pills
- On pill click: inserts into textarea

---

## Pages

| Route | Middleware | Layout | Purpose |
|-------|-----------|--------|---------|
| `/` | auth | default | Empty state; redirects to first channel |
| `/login` | guest | auth | Login form |
| `/register` | guest | auth | Registration form |
| `/profile` | auth | default | Triggers profile panel (slide-in); immediately redirects to / |
| `/channels/:id` | auth | default | Main messaging view |
| `/debug` | none | default | Dev/testing page (accessible unauthenticated) |

---

## BFF Layer (Nuxt Server Routes)

Located in `server/api/`. These are thin proxies that forward to the backend API (default `http://localhost:8000`), with special handling for auth:

| Route | Method | Purpose |
|-------|--------|---------|
| `/api/auth/login` | POST | Proxy login → set httpOnly `refresh_token` cookie |
| `/api/auth/register` | POST | Proxy register → set httpOnly cookie |
| `/api/auth/logout` | POST | Proxy logout → clear cookie |
| `/api/auth/refresh` | POST | Proxy refresh using cookie |
| `/api/channels` | GET, POST | Proxy channels endpoints |
| `/api/users/me` | GET, PATCH | Proxy profile endpoints |
| `/api/messages` | GET, POST | Proxy message endpoints |
| `/api/assist/suggestions` | POST | Proxy to assist service |

**Why a BFF?**
- Manages httpOnly cookies (cannot be accessed by JavaScript)
- Hides internal API gateway URLs (can change without frontend rebuild)
- Can add request/response transformations without touching Vue components

---

## Composables

All composables use Nuxt's `useState` for shared, SSR-safe state.

### `useAuth`

**State:** `{ user_id, access_token, expires_at }`

**Methods:**
- `register(email, password)` — Create account
- `login(email, password)` — Authenticate
- `logout()` — Revoke session
- `refreshToken()` — Get new access token
- `getProfile()` — Fetch own profile
- `validateToken()` — Check token validity

**Computed:**
- `isAuthenticated` — `user_id !== null`
- `tokenExpiresIn` — Seconds until expiry (computed reactively)

---

### `useAuthFetch`

Wraps `$fetch` with automatic token injection and retry logic:
```typescript
const response = await authFetch('/api/users/me')  // auto-adds Authorization header
```

**Retry logic:**
- On 401: call `refreshToken()`, retry once
- On second 401: clear auth state, throw error

---

### `useAssist`

**Config:**
- Debounce: 500ms
- Min length: 10 chars
- Max suggestions: 3
- Request timeout: 10s (AbortController)

**Method:**
- `suggest(channelId, input)` — Fetch suggestions (debounced)

**Returns:** Array of up to 3 strings

---

### `useChannels`

**State:** `channels: Channel[]`

**Methods:**
- `fetchMyChannels()` — List all user's channels
- `createChannel(name, description, is_private, type, initial_member_ids)` — Create
- `createDM(other_user_id)` — Start DM
- `deleteChannel(id)` — Archive
- `fetchChannel(id)` — Get details
- `fetchMembers(id)` — List members
- `addMember(id, user_id, role)` — Add member
- `removeMember(id, user_id)` — Remove member

All use `authFetch`.

---

### `useDMPeers`

**State:** `dmPeers: Record<channelId, userId>`

Maps each DM channel ID to the other user's UUID. Used to look up DM peer names in the sidebar.

---

### `useIsMobile`

**State:** `isMobile: boolean` (reactive)

Checks `window.matchMedia('(max-width: 767px)')`. Updates on `resize` event.

---

### `useLocale`

Wrapper around `@nuxtjs/i18n`:

**Methods:**
- `setLocale(locale)` — Switch language (immediately applies + saves to localStorage + PATCH to backend)

**Behavior:** Persists to `localStorage['colloquia.locale']`.

---

### `useMessageStore`

**State:** `messageStore: Record<channelId, Message[]>`

**Methods:**
- `get(channelId)` — Retrieve messages for a channel
- `setHistory(channelId, messages)` — Merge fetched history with live SSE events (merge-aware; prevents loss of live events)
- `append(message)` — Add a new message (deduplicates by ID)
- `clearChannel(channelId)` — Discard messages for a channel

**Dedup:** Messages are keyed by ID; appending a duplicate (same ID) overwrites the old entry.

---

### `useMessaging`

**Methods:**
- `fetchMessages(channelId, options)` — Fetch history with cursor pagination
  - `options.beforeId` — cursor for pagination
  - `options.limit` — default 50
- `sendMessage(channelId, content)` — Post a message

---

### `useNotifications`

**State:** `notifications: AppNotification[]` (array of `{ type, message, ... }`)

**Methods:**
- `addNotification(type, data)` — Add toast (auto-dismissed after 3.5s by default layout)
- `markAllRead()` — Clear all notifications
- `markRead(id)` — Mark single notification read
- `markChannelRead(channelId)` — Mark all mentions from a channel as read

**Computed:**
- `unreadCount` — Number of distinct unread channel IDs

---

### `usePresence`

Opens SSE stream to `/api/users/presence/stream`. Maintains `presenceMap`.

**Methods:**
- `openPresenceStream()` — Connect to SSE (called by default layout)
- `closePresenceStream()` — Disconnect

**Reconnection:** Exponential backoff (2s base, capped at 30s) on connection error.

---

### `useSSE`

Opens SSE stream to `/api/v1/messages/stream` for all subscribed channels.

**Methods:**
- `openStream(channelIds, onMessage)` — Open SSE
- `closeStream()` — Close
- `updateChannels(newChannelIds)` — Update subscription (auto-reconnects if needed)

**Reconnection:** Exponential backoff; also reconnects when channel list changes.

---

### `useTokenRefresh`

Watches `auth.expires_at`. Schedules a `refreshToken()` call 60 seconds before expiry.

---

### `useUsersCache`

**State:** `usersCache: Record<userId, displayName>`

**Methods:**
- `resolveUser(userId)` — Get user name (cached or fetched)
- `prefetchUsers(ids)` — Bulk-fetch missing IDs

**Lazy Loading:** If user not in cache, triggers background fetch.

---

## Components

### Feature Components

**`CreateChannelModal.vue`**
- Modal with form fields: name, description, is_private toggle, initial member search
- Supports typeahead search (min 2 chars, 300ms debounce)
- Navigates to new channel on success

**`ManageChannelModal.vue`**
- Tabbed modal: Members / Settings
- **Members tab**: list members with roles, add members (search + typeahead), remove members (if owner/admin)
- **Settings tab** (owner only): delete channel with two-step confirmation

**`NewDMModal.vue`**
- Search for user (min 2 chars, 300ms debounce)
- Filters out current user
- Creates DM and navigates

**`NewGroupModal.vue`**
- Shows existing DM members as read-only chips
- Add more members via search
- Optional group name
- Creates `type: group` channel

**`NotificationCenter.vue`**
- Bell icon with unread count badge
- Dropdown panel lists notifications (by type: mention, agent, message)
- "Mark all read" button

**`ProfilePanel.vue`**
- Slide-in right panel (`sm:w-80`)
- Edit fields: name, bio, language selector
- Shows avatar with initials fallback
- PATCH on save; handles 401 → refresh token → retry

**`MessageToast.vue`**
- Auto-dismissing toast stack (`fixed bottom-4 right-4`)
- Each toast shows: avatar, sender name, message preview, channel name
- Uses `TransitionGroup` for smooth enter/leave

---

### UI Primitives

**`Avatar.vue`**
- Sizes: `sm`, `md`, `lg`, `xl`
- Fallback: initials on no `src`
- Optional presence dot (green online, zinc offline)
- Presence dot only renders if `online` prop is explicitly provided

**`Button.vue`**
- Variants: `primary` (gold), `secondary`, `ghost`, `danger`
- `loading` state shows "A carregar…" + disabled
- Spreads `$attrs` for flexibility

**`Card.vue`**
- Simple wrapper: `bg-card rounded-lg border border-border p-6`

**`Input.vue`**
- Controlled input with `v-model`
- `error` prop: red border + error message below
- Password visibility toggle (eye/eye-off icons)

---

## Layouts

### `auth.vue`
- Centered single-column layout
- Heading: "Colloquia"
- Tagline: `$t('auth.tagline')`
- `<slot />` for login/register forms

### `default.vue`
The main application shell. Key responsibilities:

1. **Sidebar** — channels list, DM list, new channel buttons
2. **Unread badges** — per-channel unread count
3. **SSE lifecycle** — connect `useSSE` on mount, subscribe to all user channels
4. **Presence tracking** — `usePresence` with heartbeat
5. **DM peer resolution** — `useDMPeers` + `useUsersCache`
6. **Modal management** — state for show/hide each modal
7. **Toast queue** — display notifications
8. **Responsive sidebar** — collapsible on mobile with backdrop
9. **Provide auth-related values** — `showProfile`, `openSidebar`

---

## Middleware

### `auth.ts`
- Redirects to `/login` if `!isAuthenticated` and `authReady === true`
- Waits for `authReady` to prevent flash-redirect during initial token refresh

### `guest.ts`
- Redirects to `/` if already authenticated
- Prevents logged-in users from accessing `/login` and `/register`

---

## Plugins

### `auth.ts` (Universal)
Runs on every app load:
- If no access token in state: call `refreshToken()` (uses httpOnly cookie)
- Set `authReady = true` when done (unblocks middleware)

### `i18n-hydrate.client.ts` (Client-Only)
Resolves locale priority:
1. Authenticated user's profile `language` field
2. `localStorage['colloquia.locale']`
3. Default: `"en"`

---

## i18n

**Locales:** `en` (English), `pt` (Português)

**Strategy:** `no_prefix` (no `/en/` or `/pt/` path prefixes)

**Detection:** Disabled at browser level; resolved by `i18n-hydrate` plugin at runtime

**Scope:** All UI strings (buttons, labels, error messages, etc.)

---

## Design System

### Dark Theme
- **Background:** `hsl(0 0% 7.1%)` (near-black)
- **Sidebar:** `hsl(0 0% 8.5%)` (slightly lighter)
- **Primary / Accent:** `hsl(43 100% 50%)` (amber/gold)
- **Border:** `hsl(0 0% 20%)` (dark gray)
- **Text:** `hsl(0 0% 90%)` (off-white)

All colors are CSS custom properties (`:root`), editable in `tailwind.config.ts`.

### Typography
- **Headings:** Inter (sans-serif)
- **Body / Prose:** Source Serif 4 (serif)
- Both loaded from Google Fonts

### Animations
- `animate-fade-in` — opacity + translateY
- `animate-slide-in` — translateX from right

Defined in `tailwind.config.ts`.

### Scrollbars
Custom webkit scrollbars (6px width, transparent track, primary color thumb).

---

## Development

### Install & Run

```bash
cd apps/web
yarn install
yarn dev
```

Open http://localhost:3000.

### Build for Production

```bash
yarn build
yarn preview  # Preview the production build locally
```

### Test

```bash
yarn test                # Run all tests
yarn test:coverage       # Generate coverage report
```

Tests are in `app/composables/__tests__/` (co-located with composables).

---

## Environment Variables

Frontend-only variables (prefixed `NUXT_PUBLIC_`):

| Variable | Default | Description |
|----------|---------|-------------|
| `NUXT_PUBLIC_API_BASE` | `http://localhost:8000` | Backend API gateway URL |

Set in `.env.local` or pass at build time:

```bash
NUXT_PUBLIC_API_BASE=https://api.example.com yarn build
```

---

## Error Handling

### HTTP Errors
- **400**: Validation error (shown in toast)
- **401**: Unauthenticated (redirects to login)
- **403**: Forbidden (shown in toast)
- **409**: Conflict (shown in toast, e.g., email taken)
- **500**: Server error (shown in toast)

### SSE Disconnections
- Automatic reconnection with exponential backoff
- User sees "offline" state if connection is down for >30s
- Connection resumes automatically when network returns

---

## Performance

### Code Splitting
Nuxt automatically splits code by route and composable. Each page loads only what it needs.

### Lazy Loading
- User profiles: lazy-loaded by `useUsersCache`
- Modals: rendered conditionally (not mounted until needed)
- Components: auto-imported via Nuxt

### Memoization
- `computed` properties are cached (Vue reactivity)
- Composables are shared via `useState` (instance-level cache)

---

## Browser Support

- Modern browsers (ES2020 minimum)
- Requires SSE support (all modern browsers)
- Requires localStorage
- Mobile: iOS 13+, Android 8+

---

## Known Limitations

1. **Message ordering** — When subscribed to multiple channels via SSE, messages are interleaved (no global ordering)
2. **Typing indicators** — Not implemented (could be added via SSE events)
3. **Message reactions** — Not implemented
4. **Thread replies** — Not implemented (all messages are at the top level)
5. **Message search** — Not implemented
6. **File uploads** — Not implemented

# yarn
yarn build

# bun
bun run build
```

Locally preview production build:

```bash
# npm
npm run preview

# pnpm
pnpm preview

# yarn
yarn preview

# bun
bun run preview
```

Check out the [deployment documentation](https://nuxt.com/docs/getting-started/deployment) for more information.
