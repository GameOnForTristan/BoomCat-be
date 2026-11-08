## ADDED Requirements

### Requirement: Configuration loading
The system SHALL load configuration from YAML files with environment variable overrides. The config SHALL include Server, Redis, Database, Auth, Game, and Log sections with sensible defaults.

#### Scenario: Load config with defaults
- **WHEN** the system starts with a minimal config file
- **THEN** the system SHALL apply default values for missing configuration fields

### Requirement: Structured logging
The system SHALL provide a structured logger with configurable levels (debug, info, warn, error), output formats (json, text), and output targets (stdout, file). The logger SHALL support contextual fields.

#### Scenario: Log with context
- **WHEN** a service logs a game event
- **THEN** the logger SHALL include contextual fields like room_id and player_id in the output

### Requirement: Error codes and types
The system SHALL define domain-specific error types and error codes. HTTP errors SHALL use the unified error code system (100x for auth, 200x for room, 300x for player, 400x for game, 500x for server).

#### Scenario: Return domain error
- **WHEN** a room operation fails because the room is full
- **THEN** the system SHALL return error code 2002 with message "room_full"

### Requirement: Utility functions
The system SHALL provide utility functions for: generating unique IDs (user IDs with u_ prefix, room IDs with r_ prefix, card IDs), generating 6-character invite codes, and time-related helpers.

#### Scenario: Generate user ID
- **WHEN** a new user registers
- **THEN** the system SHALL generate a unique ID with format "u_" + random string