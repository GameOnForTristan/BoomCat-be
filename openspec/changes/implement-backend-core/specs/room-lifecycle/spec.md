## ADDED Requirements

### Requirement: Room creation and management
The system SHALL support creating a room with a unique ID and 6-character invite code. The room SHALL have configurable max players (2-5), owner, and lifecycle status (waiting, playing, finished).

#### Scenario: Create a room
- **WHEN** a player creates a room with max_players=4
- **THEN** the system SHALL return a room with a unique room_id, 6-character invite_code, status="waiting", owner_id set to the creator

### Requirement: Player room lifecycle
The system SHALL support players joining a room by invite code, leaving a room, toggling ready status, and starting the game (owner only, all players must be ready).

#### Scenario: Join room by invite code
- **WHEN** a player sends a valid invite_code
- **THEN** the system SHALL add the player to the room and return the full room state

#### Scenario: Start game without all ready
- **WHEN** the owner tries to start the game but not all players are ready
- **THEN** the system SHALL reject with error code "not_all_ready"

### Requirement: Player connection status
The system SHALL track each player's WebSocket connection status: Connected, Disconnected, TimedOut. Disconnected players SHALL have a 90-second reconnection window.

#### Scenario: Player disconnects
- **WHEN** a player's WebSocket connection drops
- **THEN** the system SHALL mark the player as Disconnected and start a 90-second reconnection timer