# Contributing to Railpack

## Project Status

This is an early-stage project that is expected to undergo frequent changes. While we welcome contributions, please note that the API and functionality may change significantly as we evolve.

## Pull Requests

We welcome pull requests that push the project forward in meaningful ways. Please ensure your PRs:

- Address a specific problem or add a well-defined feature
- Include tests for new functionality
- Follow the existing code style

Note: We prefer focused, well-thought-out contributions over "drive-by" PRs that make superficial changes.

## Testing

### Core Tests

- All example plans are snapshot tested in `core_test.go`
- Tests with a `test.json` file will be built and run automatically
- The test output must contain the `expectedOutput` specified in the test file

## Useful Commands

- `mise check` - Run linting and type checking
- `mise test` - Run unit tests
- `mise test-integration` - Run integration tests
