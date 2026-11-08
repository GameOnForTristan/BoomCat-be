## ADDED Requirements

### Requirement: User registration
The system SHALL support user registration with nickname (2-16 characters) and password (6-32 characters). The password SHALL be hashed with bcrypt before storage. On success, the system SHALL return a user_id and JWT token.

#### Scenario: Register a new user
- **WHEN** a user registers with a valid nickname and password
- **THEN** the system SHALL create a user record with bcrypt-hashed password and return user_id + JWT token

### Requirement: User login
The system SHALL support user login with user_id and password. The system SHALL verify the password against the stored bcrypt hash. On success, the system SHALL return a new JWT token with expiry.

#### Scenario: Login with wrong password
- **WHEN** a user logs in with an incorrect password
- **THEN** the system SHALL return an "unauthorized" error

### Requirement: JWT token management
The system SHALL issue JWT tokens with configurable expiry. The system SHALL support token refresh. The system SHALL cache active sessions in Redis for fast authentication.

#### Scenario: Token refresh
- **WHEN** a user presents a valid, non-expired token to the refresh endpoint
- **THEN** the system SHALL return a new JWT token with extended expiry