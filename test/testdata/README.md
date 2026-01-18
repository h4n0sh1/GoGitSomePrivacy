# Sample test data for GoGitSomePrivacy

This directory contains test data for unit tests and integration tests.

## Structure

- `commits/`: Sample commit data
- `repos/`: Sample repository metadata
- `expected/`: Expected output for various test scenarios

## Usage

Test files can reference data in this directory using relative paths from the test file location.

Example:
```go
data, err := os.ReadFile("../../test/testdata/commits/sample.json")
```
