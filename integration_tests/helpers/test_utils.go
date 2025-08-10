package helpers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	BinaryPath    string
	TempDir       string
	BaseDeckPath  string
	ConfigFile    string
}

// NewTestConfig creates a new test configuration with temporary directories
func NewTestConfig() (*TestConfig, error) {
	tempDir, err := os.MkdirTemp("", "srs_integration_test_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	baseDeckPath := filepath.Join(tempDir, "test_decks")
	if err := os.MkdirAll(baseDeckPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base deck path: %v", err)
	}

	// Create SRS config directory structure
	srsConfigDir := filepath.Join(tempDir, "srs")
	if err := os.MkdirAll(srsConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create srs config dir: %v", err)
	}
	
	configFile := filepath.Join(srsConfigDir, "config")
	configContent := fmt.Sprintf("# SRS Test Configuration\nbase_deck=%s\n", baseDeckPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %v", err)
	}

	// Find the SRS binary - prefer test binary if it exists
	// Try different path levels depending on where we're running from
	var binaryPath string
	candidates := []string{"../../srs-test", "../srs-test", "./srs-test", "../../srs", "../srs", "./srs"}
	
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			binaryPath = candidate
			break
		}
	}
	
	if binaryPath == "" {
		return nil, fmt.Errorf("SRS binary not found. Please build with: go build -o srs .")
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for binary: %v", err)
	}
	binaryPath = absPath

	return &TestConfig{
		BinaryPath:   binaryPath,
		TempDir:      tempDir,
		BaseDeckPath: baseDeckPath,
		ConfigFile:   configFile,
	}, nil
}

// Cleanup removes temporary test files
func (tc *TestConfig) Cleanup() error {
	return os.RemoveAll(tc.TempDir)
}

// RunCommand executes an SRS command with the test configuration
func (tc *TestConfig) RunCommand(args ...string) (*CommandResult, error) {
	cmd := exec.Command(tc.BinaryPath, args...)
	
	// Set environment to use our test config directory
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+tc.TempDir)
	
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)
	
	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Duration: duration,
	}
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		} else {
			return nil, fmt.Errorf("failed to run command: %v", err)
		}
	}
	
	return result, nil
}

// CommandResult holds the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Success returns true if the command succeeded
func (cr *CommandResult) Success() bool {
	return cr.ExitCode == 0
}

// MCPClient provides a simple client for testing MCP server functionality
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	stderr *bufio.Scanner
}

// MCPRequest represents an MCP JSON-RPC request
type MCPRequest struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP JSON-RPC response
type MCPResponse struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPClient starts an MCP server and returns a client to interact with it
func (tc *TestConfig) NewMCPClient() (*MCPClient, error) {
	cmd := exec.Command(tc.BinaryPath, "mcp")
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+tc.TempDir)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %v", err)
	}
	
	client := &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
		stderr: bufio.NewScanner(stderr),
	}
	
	// Wait for initialization response
	time.Sleep(100 * time.Millisecond)
	
	return client, nil
}

// SendRequest sends a JSON-RPC request and waits for response
func (mc *MCPClient) SendRequest(method string, params interface{}) (*MCPResponse, error) {
	// Use unique request ID
	requestID := int(time.Now().UnixNano() % 1000000)
	request := MCPRequest{
		ID:     requestID,
		Method: method,
		Params: params,
	}
	
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	if _, err := mc.stdin.Write(append(requestBytes, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %v", err)
	}
	
	// Read response with timeout and ID matching
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("request timed out")
		default:
			if mc.stdout.Scan() {
				line := mc.stdout.Text()
				if line == "" {
					continue
				}
				
				var response MCPResponse
				if err := json.Unmarshal([]byte(line), &response); err != nil {
					// Skip non-JSON lines (like initialization)
					continue
				}
				
				// Check if this is the response to our request
				if response.ID == requestID {
					return &response, nil
				}
				// Otherwise keep looking for the right response
			}
		}
	}
}

// Close terminates the MCP server
func (mc *MCPClient) Close() error {
	if mc.stdin != nil {
		mc.stdin.Close()
	}
	if mc.cmd != nil {
		return mc.cmd.Process.Kill()
	}
	return nil
}

// LogResult logs test results to a file
func LogResult(testName string, result *CommandResult, logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}
	
	logFile := filepath.Join(logDir, testName+".log")
	logContent := fmt.Sprintf("Test: %s\nExit Code: %d\nDuration: %v\n\nSTDOUT:\n%s\n\nSTDERR:\n%s\n",
		testName, result.ExitCode, result.Duration, result.Stdout, result.Stderr)
	
	return os.WriteFile(logFile, []byte(logContent), 0644)
}