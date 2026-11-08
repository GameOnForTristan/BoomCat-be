## ADDED Requirements

### Requirement: Game service orchestration
The game service SHALL coordinate the complete play_card flow: validate move legality, update hand/deck/discard, push to action stack, open Nope window, and trigger state broadcast. The service SHALL act as the bridge between Controller and Model layers.

#### Scenario: Complete play card flow
- **WHEN** a player plays a valid card combination
- **THEN** the game service SHALL validate, execute, open Nope window, and trigger async state broadcast

### Requirement: Turn service scheduling
The turn service SHALL manage turn transitions, including direction changes, player elimination removal from rotation, and attack debt accumulation. The service SHALL enforce that only the current player can take actions.

#### Scenario: Non-current player tries to act
- **WHEN** a player who is not the current turn tries to play a card
- **THEN** the turn service SHALL reject the action with "not your turn" error

### Requirement: Timer service management
The timer service SHALL manage three types of timers: turn timeout (60s default), Nope window (3s), and reconnection timeout (90s). The service SHALL support resetting the Nope timer when a new Nope is played.

#### Scenario: Turn timeout
- **WHEN** the current player takes no action for 60 seconds
- **THEN** the timer service SHALL trigger an automatic draw action for that player

### Requirement: Room service operations
The room service SHALL coordinate room creation, joining, leaving, ready toggle, and game start. The service SHALL validate that the game can only start when all players are ready and the player count is at least 2.

#### Scenario: Start game with minimum players
- **WHEN** 2 players are in the room and both are ready
- **THEN** the room service SHALL start the game and transition the room status to "playing"