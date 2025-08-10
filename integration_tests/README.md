# SRS Integration Tests

This directory contains comprehensive integration tests for the SRS (Spaced Repetition System) binary.

## Test Coverage

### CLI Command Tests (`cli_tests/`)
- **Basic Commands**: `version`, `list`, `config`, help functionality
- **Deck Operations**: Listing decks, handling invalid paths, performance testing
- **Error Handling**: Invalid commands, non-existent paths, malformed input

### MCP Server Tests (`mcp_tests/`)
- **Server Lifecycle**: Startup, shutdown, basic connectivity
- **Tool Operations**: 
  - `srs/get_due_cards` - Retrieve cards ready for review
  - `srs/rate_card` - Update card scheduling based on review performance
  - `srs/get_deck_stats` - Get deck statistics and metrics
  - `srs/list_decks` - List all available decks
- **Error Scenarios**: Invalid tools, malformed parameters, non-existent resources
- **Performance**: Response times, concurrent request handling

### End-to-End Workflow Tests (`workflow_test.go`)
- **Complete Card Lifecycle**: Create → List → Review → Verify scheduling updates
- **Multi-Deck Operations**: Cross-deck operations and management
- **Review Progression**: Multiple review sessions with proper FSRS scheduling
- **Error Recovery**: System behavior under various failure conditions
- **Performance Testing**: Large dataset handling and response times

## Test Infrastructure

### Fixtures (`fixtures/`)
- **Test Cards**: Pre-defined card sets for different scenarios
- **Test Decks**: Various deck configurations (basic math, programming, reviewed cards)
- **FSRS Metadata**: Cards with and without existing review history

### Helpers (`helpers/`)
- **Test Configuration**: Isolated test environments with temporary directories
- **Command Execution**: Wrapper for running SRS binary with proper configuration
- **MCP Client**: JSON-RPC client for testing MCP server functionality
- **Logging**: Test result logging and artifact collection

## Running Tests

### Manual Execution
```bash
# Run all integration tests
./integration_tests/run_tests.sh

# Run specific test categories
cd integration_tests
go test -v ./cli_tests/
go test -v ./mcp_tests/
go test -v ./workflow_test.go
```

### CI Integration
Tests run automatically in GitHub Actions:
- **Unit Tests**: Standard Go tests (`go test ./...`)
- **Integration Tests**: Binary testing with real filesystem operations
- **Artifact Collection**: Test logs uploaded for debugging failed runs

## Test Environment

Each test creates an isolated environment:
- **Temporary Directories**: Clean filesystem for each test run
- **Isolated Configuration**: Test-specific config files
- **Binary Detection**: Automatically finds `srs-test` or `srs` binary
- **Cleanup**: Automatic cleanup of test artifacts

## What's NOT Tested

- **TUI Interactions**: Terminal UI is complex to test reliably
- **Interactive Commands**: Commands requiring user input
- **File System Permissions**: Platform-specific permission handling
- **Network Operations**: Update checks and remote operations

## Test Philosophy

These integration tests focus on:
1. **Real Binary Testing**: Tests the actual compiled binary, not just code
2. **End-to-End Workflows**: Complete user scenarios from start to finish  
3. **MCP Protocol Compliance**: Proper JSON-RPC behavior and error handling
4. **Performance Validation**: Reasonable response times under normal loads
5. **Error Resilience**: Graceful handling of invalid input and edge cases

## Adding New Tests

1. **CLI Tests**: Add to `cli_tests/basic_commands_test.go`
2. **MCP Tests**: Add to `mcp_tests/mcp_server_test.go`
3. **Workflow Tests**: Add to `workflow_test.go`
4. **Fixtures**: Add new test data to `fixtures/test_cards.go`
5. **Helpers**: Add utilities to `helpers/test_utils.go`

Follow existing patterns for test isolation and cleanup.