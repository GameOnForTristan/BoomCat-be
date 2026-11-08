## ADDED Requirements

### Requirement: Unified API response format
All HTTP API responses SHALL use the format `{"code": 0, "message": "success", "data": {...}}`. Error responses SHALL use the same format with non-zero code and descriptive message.

#### Scenario: Successful API response
- **WHEN** an API call succeeds
- **THEN** the response SHALL have code=0, message="success", and data containing the result

### Requirement: Auth API endpoints
The system SHALL expose POST `/api/v1/auth/register`, POST `/api/v1/auth/login`, and POST `/api/v1/auth/refresh` endpoints. The register endpoint SHALL accept nickname and password. The login endpoint SHALL accept user_id and password.

#### Scenario: Register endpoint
- **WHEN** a POST request is sent to `/api/v1/auth/register` with valid nickname and password
- **THEN** the system SHALL return 200 with user_id and token in the data field

### Requirement: Room API endpoints
The system SHALL expose POST `/api/v1/rooms` (create), GET `/api/v1/rooms/{room_id}` (detail), POST `/api/v1/rooms/join` (join by invite code), POST `/api/v1/rooms/{room_id}/leave` (leave), POST `/api/v1/rooms/{room_id}/ready` (toggle ready), and POST `/api/v1/rooms/{room_id}/start` (start game, owner only).

#### Scenario: Create room endpoint
- **WHEN** a POST request is sent to `/api/v1/rooms` with optional name and max_players
- **THEN** the system SHALL create a room and return room_id, invite_code, name, max_players, and status