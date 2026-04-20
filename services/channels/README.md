# Channels Service

**Channels, direct messages, groups, and membership management.**

The Channels service manages the container structure for conversations: public channels, private channels, direct message (DM) conversations, and group chats. It enforces access control (who can read/write), manages membership with role-based permissions, and handles channel lifecycle.

---

## What It Does

- **Channel Management** — Create, retrieve, delete (archive) channels
- **Channel Types** — Support public channels, private channels, DMs, and groups
- **Membership** — Add/remove members, manage roles (owner/admin/member)
- **Access Control** — Enforce private channel visibility, role-based actions
- **DM Idempotency** — Guarantee each pair of users has at most one DM conversation
- **Membership Validation** — gRPC RPC for messaging service to verify sender membership

---

## Channel Types

| Type | Archived? | Private? | Roles | Rules |
|------|-----------|----------|-------|-------|
| **`channel`** | Soft-delete | User-defined | owner/admin/member | Public by default; name required; members by invite |
| **`group`** | Soft-delete | Always private | owner/admin/member | Private upgrade from DM; group name optional |
| **`dm`** | Cannot delete | N/A | N/A | Two users only; idempotent via SHA-256 key; no membership changes |

### DM Idempotency

When creating a DM between users A and B:
1. Compute `dm_key = SHA256(sort([A_id, B_id]))`
2. Check if a DM with this key already exists
3. If yes: return existing channel with `created=false` (already existed)
4. If no: create new channel with `created=true`

This ensures that calling `CreateDM` twice with the same users always returns the same channel.

### Group Behavior

A group channel:
- Is always private (`is_private = true`)
- Can have 2+ members (unlike a pure DM)
- Can be upgraded from a DM conversation
- Supports optional `name` field (unlike DMswhich have no name)

---

## HTTP API

All endpoints require JWT authentication (Bearer token).

### Channels

#### POST `/api/v1/channels`

Create a new channel.

**Request Body:**
```json
{
  "name": "project-alpha",
  "description": "Discussion of project alpha",
  "is_private": false,
  "type": "channel",
  "initial_member_ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "550e8400-e29b-41d4-a716-446655440001"
  ]
}
```

**Constraints:**
- `name` — max 80 characters; required for `channel` and `group` types; ignored for `dm`
- `description` — max 500 characters
- `is_private` — boolean; ignored for `group` (always private)
- `type` — `"channel"`, `"group"`, or `"dm"`; `dm` cannot be created via this endpoint (use `/channels/dm`)
- `initial_member_ids` — UUIDs to add at creation time; creator is always added as owner

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "project-alpha",
  "description": "Discussion of project alpha",
  "is_private": false,
  "type": "channel",
  "created_by": "550e8400-e29b-41d4-a716-446655440000",
  "archived": false,
  "member_count": 3,
  "created_at": 1639747350,
  "updated_at": 1639747350
}
```

**Error Responses:**
- `400 Bad Request` — Validation error (invalid type, missing name, etc.)
- `409 Conflict` — Channel name already taken (within user's scope)

---

#### POST `/api/v1/channels/dm`

Create or retrieve a DM channel between two users.

**Request Body:**
```json
{
  "other_user_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Response (201 or 200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "name": null,
  "type": "dm",
  "created_by": null,
  "archived": false,
  "member_count": 2,
  "created": true,    // true if newly created; false if already existed
  "created_at": 1639747350,
  "updated_at": 1639747350
}
```

---

#### GET `/api/v1/channels/me`

List all channels the authenticated user is a member of.

**Response (200 OK):**
```json
{
  "channels": [
    { "id": "...", "name": "...", "type": "...", ... },
    ...
  ]
}
```

---

#### GET `/api/v1/channels/:id`

Retrieve a single channel by ID (only if you are a member; private channels are not visible to non-members).

**Response (200 OK):** Channel object (same as POST response).

**Error Responses:**
- `404 Not Found` — Channel not found
- `403 Forbidden` — You are not a member (for private channels)

---

#### DELETE `/api/v1/channels/:id`

Archive (soft-delete) a channel. Only the owner or an admin may do this.

**Response (204 No Content):** Empty body.

**Error Responses:**
- `404 Not Found` — Channel not found
- `403 Forbidden` — You are not the owner/admin
- `422 Unprocessable Entity` — Channel is a DM (cannot delete DMs)

---

### Membership

#### GET `/api/v1/channels/:id/members`

List all members of a channel.

**Response (200 OK):**
```json
{
  "members": [
    {
      "channel_id": "550e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "owner",
      "joined_at": 1639747350
    },
    ...
  ]
}
```

---

#### POST `/api/v1/channels/:id/members`

Add a member to a channel.

**Request Body:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "role": "member"
}
```

**Constraints:**
- `role` — `"owner"`, `"admin"`, or `"member"`
- For **private channels**: only the owner/admin of that channel may add members
- For **public channels**: any member may add members (optional; enforce server-side)
- **DM channels**: cannot add members (400 error)
- **Archived channels**: cannot add members (422 error)

**Response (201 Created):** Member object.

**Error Responses:**
- `403 Forbidden` — You are not the owner/admin (for private channels)
- `422 Unprocessable Entity` — Channel is a DM or archived
- `409 Conflict` — User already a member

---

#### DELETE `/api/v1/channels/:id/members/:userId`

Remove a member from a channel. Members can remove themselves; non-owners can only remove themselves.

**Response (204 No Content):** Empty body.

**Error Responses:**
- `403 Forbidden` — You are not the owner/admin (and are trying to remove someone else)
- `404 Not Found` — Member not found

---

## gRPC API

The gRPC service runs on port 50053. RPCs are defined in `proto/channels.proto`.

| RPC | Request | Response |
|-----|---------|----------|
| `CreateChannel` | `name`, `description`, `is_private`, `type`, `initial_member_ids[]` | `Channel` |
| `CreateDM` | `other_user_id` | `Channel` |
| `GetChannel` | `id` | `Channel` |
| `DeleteChannel` | `id` | `DeleteChannelResponse` (empty) |
| `AddMember` | `channel_id`, `user_id`, `role` | `Member` |
| `RemoveMember` | `channel_id`, `user_id` | `RemoveMemberResponse` (empty) |
| `ListUserChannels` | `user_id` | `ListUserChannelsResponse` (`channels[]`) |
| `ListChannelMembers` | `channel_id` | `ListChannelMembersResponse` (`members[]`) |
| `ValidateMembership` | `channel_id`, `user_id` | `ValidateMembershipResponse` (`is_member: bool`) |

The **`ValidateMembership`** RPC is called by the messaging service before allowing a message to be sent.

---

## Database Schema

### `channels` Table

```sql
CREATE TABLE channels (
    id UUID PRIMARY KEY,
    name TEXT,                          -- max 80; nullable for DMs
    description TEXT,                   -- max 500
    is_private BOOLEAN DEFAULT FALSE,
    created_by UUID,                    -- who created the channel
    archived BOOLEAN DEFAULT FALSE,
    type TEXT NOT NULL,                 -- 'channel', 'dm', 'group'
    dm_key TEXT UNIQUE,                 -- SHA-256 of sorted user IDs for DMs; nullable
    member_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_is_private ON channels(is_private);
CREATE INDEX idx_channels_created_by ON channels(created_by);
CREATE UNIQUE INDEX idx_channels_dm_key_partial ON channels(dm_key) WHERE type='dm';
```

### `channel_members` Table

```sql
CREATE TABLE channel_members (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',  -- 'owner', 'admin', 'member'
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX idx_channel_members_user_id ON channel_members(user_id);
CREATE INDEX idx_channel_members_channel_id ON channel_members(channel_id);
```

---

## Environment Variables

All required at startup.

| Variable | Description | Example |
|----------|-------------|---------|
| `CHANNELS_DATABASE_URL` | PostgreSQL DSN | `postgres://postgres:password@localhost:5436/channels?sslmode=disable` |
| `JWT_PUBLIC_KEY` | RS256 public key (PEM format) | (multi-line PEM content) |
| `CHANNELS_GRPC_PORT` | Port for gRPC server | `50053` |
| `CHANNELS_HTTP_PORT` | Port for HTTP server | `8083` |
| `USERS_GRPC_ADDRESS` | Address of users service (best-effort validation) | `localhost:50052` |

---

## Development

### Run Migrations

```bash
cd services/channels
export DATABASE_URL="postgres://postgres:password@localhost:5436/channels?sslmode=disable"
make migrate-up
```

### Run Service Locally

```bash
export CHANNELS_DATABASE_URL="postgres://postgres:password@localhost:5436/channels?sslmode=disable"
export JWT_PUBLIC_KEY="$(cat ../../dev/public.pem)"
export CHANNELS_GRPC_PORT=50053
export CHANNELS_HTTP_PORT=8083
export USERS_GRPC_ADDRESS="localhost:50052"

make run
```

### Run Tests

```bash
cd services/channels
make test
```

---

## Role-Based Access Control

| Action | Owner | Admin | Member |
|--------|-------|-------|--------|
| Add member | ✓ | ✓ | ✗ |
| Remove member | ✓ | ✓ | (self only) |
| Delete channel | ✓ | ✓ | ✗ |
| Send message | ✓ | ✓ | ✓ |
| View members | ✓ | ✓ | ✓ |

---

## Error Handling & Status Codes

| Error | HTTP Status | gRPC Code | Reason |
|-------|-------------|-----------|--------|
| `ErrChannelNotFound` | 404 | `NotFound` | Channel does not exist |
| `ErrChannelAlreadyExists` | 409 | `AlreadyExists` | Channel name already taken |
| `ErrMemberNotFound` | 404 | `NotFound` | Member not in channel |
| `ErrAccessDenied` | 403 | `PermissionDenied` | Insufficient permissions |
| `ErrDMModificationNotAllowed` | 422 | `FailedPrecondition` | Cannot modify a DM channel |
| `ErrChannelArchived` | 422 | `FailedPrecondition` | Channel is archived |

---

## Inter-Service Communication

**Called By:**
- Auth service: none
- Users service: none
- Messaging service: `ValidateMembership` RPC to check if sender is a channel member (fail-closed)

**Calls:**
- Users service: validates that added members exist (best-effort; service continues if users is down)

---

## Archived Channels

Channels are **soft-deleted** (archived), not removed from the database:
- Archived channels do not appear in `ListUserChannels`
- Archived channels cannot have members added/removed
- Old messages in archived channels are retained
- Re-opening an archived channel is not currently supported (can be added if needed)

---

## Membership Immutability

Member roles are **not updatable** once a member is added (no `UpdateMember` RPC). To change someone's role, remove and re-add them with the new role.
