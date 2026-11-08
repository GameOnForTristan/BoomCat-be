## ADDED Requirements

### Requirement: Card types and values
The system SHALL define all 9 card types: ExplodingKitten, Defuse, Skip, Attack, Favor, Shuffle, SeeTheFuture, Nope, CatCard. Each card SHALL have a unique ID, type, name, and description.

#### Scenario: Create a Defuse card
- **WHEN** a new Defuse card is created with ID "defuse_01"
- **THEN** the card SHALL have Type=defuse, Name="拆解", Description="抵消爆炸猫"

### Requirement: Deck initialization and manipulation
The system SHALL support initializing a deck with the correct card distribution for N players (N-1 ExplodingKitten, 2 Defuse, 4 Skip, 4 Attack, 4 Favor, 4 Shuffle, 5 SeeTheFuture, 5 Nope, 20 CatCard). The deck SHALL support shuffle, draw from top, insert at specific position, and peek at top N cards.

#### Scenario: Draw from deck
- **WHEN** a player draws from the top of the deck
- **THEN** the top card SHALL be removed from the deck and returned to the caller

### Requirement: Hand operations and combo validation
The system SHALL support adding/removing cards from a hand. The system SHALL validate combo moves: 3 identical CatCards to steal a specific card, 5 different CatCards to take from discard pile.

#### Scenario: Three identical CatCards combo
- **WHEN** a player plays 3 cards all of type CatCard with the same name
- **THEN** the system SHALL validate the combo as valid and remove the 3 cards from hand

### Requirement: Turn engine with attack stacking
The system SHALL maintain a circular turn order with direction (1=clockwise, -1=counterclockwise). The system SHALL track pendingDraws counter that accumulates when Attack is played and decrements when cards are drawn.

#### Scenario: Attack stacking
- **WHEN** player A plays Attack, then player B plays Attack without drawing
- **THEN** pendingDraws SHALL be 4 and player C SHALL be the next to draw

### Requirement: Action stack for Nope chain
The system SHALL maintain an action stack where each action can be Noped. Nope resolution SHALL use odd-even counting: odd number of Nopes blocks the action, even number allows it.

#### Scenario: Nope chain cancellation
- **WHEN** player A plays cards, player B plays Nope, player A plays Nope
- **THEN** the original action SHALL be unblocked (2 Nopes = even)

### Requirement: Game state machine
The system SHALL maintain game phases: Waiting, Dealing, Playing, GameOver. The system SHALL validate that operations are only allowed in the correct phase.

#### Scenario: Play card during Waiting
- **WHEN** a player tries to play a card during the Waiting phase
- **THEN** the system SHALL reject the operation with an invalid phase error

### Requirement: Game over detection
The system SHALL detect when only one player remains alive and declare that player the winner.

#### Scenario: Last player standing
- **WHEN** a player is eliminated and only one player remains
- **THEN** the game SHALL transition to GameOver phase and declare the surviving player as winner