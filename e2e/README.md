# dSock E2E Tests

The dSock end-to-end tests, written in Go (using `testing`).

## Notes

User/sessions are related to the current test running. This is to prevent test collisions.

The user is the name of the test file.

The sessions is the name of the test. If no session is set, add the name in the user after a `_`.

## Usage

Run `script/run-e2e` (file in root of project, can be ran from anywhere) to run the E2E tests.

Tests are ran inside Docker Compose, using an additional configuration (`docker-compose.e2e.yml` inside this directory).
This adds the `e2e` service on top of the normal development services.

The script also runs `docker-compose down` before running to stop all services, which clears Redis' storage.
This ensures that all tests are ran on a clean database. Redis isn't configured to persist in development.
