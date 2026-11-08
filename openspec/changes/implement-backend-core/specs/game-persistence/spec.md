## ADDED Requirements

### Requirement: Redis game state snapshot
The system SHALL serialize the full game state to Redis after each state change. The key pattern SHALL be `game_state:room:{room_id}` with TTL of 86400 seconds. The write SHALL be asynchronous and non-blocking to the game main loop.

#### Scenario: State snapshot after card play
- **WHEN** a player plays a card and the game state changes
- **THEN** the system SHALL asynchronously write the full game state JSON to Redis key `game_state:room:{room_id}`

### Requirement: Redis room metadata
The system SHALL store room metadata in Redis Hash with key pattern `room:{room_id}`, containing fields: name, owner_id, status, max_players, player_count, created_at.

#### Scenario: Create room metadata
- **WHEN** a room is created
- **THEN** the system SHALL HSET the room metadata fields into Redis

### Requirement: Redis invite code index
The system SHALL maintain an invite code to room ID mapping in Redis with key pattern `invite_code:{code}` and TTL of 86400 seconds.

#### Scenario: Lookup room by invite code
- **WHEN** a player joins with an invite code
- **THEN** the system SHALL look up the room_id from Redis key `invite_code:{code}`

### Requirement: PostgreSQL game records
The system SHALL write game records to the `boomcat_game_records` table after a game finishes. The record SHALL include players JSON, action_log JSON, and result JSON.

#### Scenario: Save game record after game over
- **WHEN** a game finishes
- **THEN** the system SHALL INSERT a record into `boomcat_game_records` with the full game data

### Requirement: PostgreSQL player stats
The system SHALL update player statistics in the `player_stats` table after each game. The stats SHALL include total_games, wins, losses, and total_kills.

#### Scenario: Update winner stats
- **WHEN** a player wins a game
- **THEN** the system SHALL increment the player's total_games and wins in `player_stats`