# Contributing to Railpack

_Note: Contributions are not being accepted at this time as we are still
focusing on setting up the core architecture._

## Project Status

This is an early-stage project that is expected to undergo frequent changes.

## Testing

### Core Tests

- All example plans are snapshot tested in `core_test.go`
- Tests with a `test.json` file will be built and run automatically
- The test output must contain the `expectedOutput` specified in the test file

## Useful Commands

- `mise check` - Run linting and type checking
- `mise test` - Run unit tests
- `mise test-integration` - Run integration tests
