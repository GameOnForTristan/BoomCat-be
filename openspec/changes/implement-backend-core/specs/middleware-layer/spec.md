## ADDED Requirements

### Requirement: Authentication middleware
The middleware SHALL extract the JWT token from the Authorization header (Bearer scheme), validate the token signature and expiry, and inject the user identity into the request context. Requests without valid tokens SHALL be rejected with HTTP 401.

#### Scenario: Missing auth header
- **WHEN** a request is made to a protected endpoint without an Authorization header
- **THEN** the middleware SHALL return HTTP 401 with error code "unauthorized"

### Requirement: CORS middleware
The middleware SHALL set Cross-Origin Resource Sharing headers to allow frontend requests from different origins during development.

#### Scenario: CORS preflight request
- **WHEN** a browser sends an OPTIONS preflight request
- **THEN** the middleware SHALL respond with appropriate CORS headers

### Requirement: Recovery middleware
The middleware SHALL recover from panics in HTTP handlers, log the stack trace, and return a 500 response with a generic error message to avoid leaking internal details.

#### Scenario: Handler panics
- **WHEN** an HTTP handler panics during request processing
- **THEN** the middleware SHALL recover, log the error, and return HTTP 500

### Requirement: Logging middleware
The middleware SHALL log each HTTP request with method, path, status code, duration, and client IP. The log output SHALL use structured JSON format.

#### Scenario: Request logging
- **WHEN** an HTTP request is processed
- **THEN** the middleware SHALL log the request details after the response is sent