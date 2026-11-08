## ADDED Requirements

### Requirement: WebSocket connection establishment
The system SHALL accept WebSocket connections at `/ws?room_id={room_id}&token={token}`. The system SHALL validate the JWT token before upgrading the connection. The system SHALL register the connection with the corresponding room.

#### Scenario: Connect with invalid token
- **WHEN** a client connects with an invalid or expired JWT token
- **THEN** the system SHALL reject the WebSocket upgrade with HTTP 401

### Requirement: Message serialization
The system SHALL use a unified JSON message format for all WebSocket communication. Client messages SHALL have `type`, `payload`, and optional `req_id` fields. Server messages SHALL have `type`, `payload`, optional `req_id`, and `timestamp` fields.

#### Scenario: Client sends action message
- **WHEN** a client sends JSON `{"type":"play_card","payload":{"card_ids":["cat_01"]}}`
- **THEN** the system SHALL parse the message and route it to the game handler

### Requirement: State broadcast with view filtering
The system SHALL broadcast state updates to all players after each state change. Each player SHALL receive a filtered view: only their own hand cards, opponents' hand sizes, deck count, and public game state.

#### Scenario: Broadcast state update
- **WHEN** the game state changes after a card play
- **THEN** the system SHALL send a `state_update` event to each player with their respective filtered view

### Requirement: Disconnection and reconnection
The system SHALL detect WebSocket disconnections. The system SHALL support reconnection within the configured timeout window (default 90 seconds). On reconnection, the system SHALL send the full current state view to the player.

#### Scenario: Player reconnects
- **WHEN** a disconnected player reconnects within 90 seconds
- **THEN** the system SHALL restore the WebSocket connection and send the full current state for that player