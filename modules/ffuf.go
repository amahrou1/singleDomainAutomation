package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"subdomain-recon/config"
	"subdomain-recon/utils"
)

// FfufResult represents the JSON output structure from ffuf
type FfufResult struct {
	Results []struct {
		URL        string `json:"url"`
		StatusCode int    `json:"status"`
	} `json:"results"`
}

// Ffuf represents the ffuf fuzzing module
type Ffuf struct {
	config     *config.Config
	ffufConfig *config.FfufConfig
}

// NewFfuf creates a new ffuf module instance
func NewFfuf(cfg *config.Config) *Ffuf {
	return &Ffuf{
		config:     cfg,
		ffufConfig: cfg.NewFfufConfig(),
	}
}

// Run executes the ffuf module
func (ff *Ffuf) Run() error {
	fmt.Println(strings.Repeat("=", 60))
	utils.InfoLogger.Printf("Starting Module: %s", ff.GetName())
	fmt.Println(strings.Repeat("=", 60))

	// Check if ffuf is installed
	if _, err := exec.LookPath("ffuf"); err != nil {
		utils.WarningLogger.Printf("Module '%s' skipped: ffuf is not installed or not in PATH", ff.GetName())
		fmt.Println(strings.Repeat("=", 60))
		return fmt.Errorf("ffuf is not installed or not in PATH")
	}

	// Check if wordlist exists
	if _, err := os.Stat(ff.ffufConfig.Wordlist); os.IsNotExist(err) {
		utils.WarningLogger.Printf("Module '%s' skipped: wordlist not found at %s", ff.GetName(), ff.ffufConfig.Wordlist)
		fmt.Println(strings.Repeat("=", 60))
		return fmt.Errorf("wordlist not found: %s", ff.ffufConfig.Wordlist)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(ff.ffufConfig.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Run ffuf
	if err := ff.runFfuf(); err != nil {
		return err
	}

	return nil
}

// runFfuf executes the ffuf command
func (ff *Ffuf) runFfuf() error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ff.config.Timeout)
	defer cancel()

	// Write target header to output file
	ff.writeTargetHeader()

	// Normalize target URL (add scheme, remove trailing slash)
	normalizedTarget := config.NormalizeTargetURL(ff.config.Target)

	// Create temp output file for this target
	tempFile := ff.ffufConfig.OutputFile + ".temp"
	defer os.Remove(tempFile) // Clean up temp file

	// Build extensions with leading dots
	var extensionArg string
	if len(ff.ffufConfig.Extensions) > 0 {
		extList := make([]string, len(ff.ffufConfig.Extensions))
		for i, ext := range ff.ffufConfig.Extensions {
			if !strings.HasPrefix(ext, ".") {
				extList[i] = "." + ext
			} else {
				extList[i] = ext
			}
		}
		extensionArg = strings.Join(extList, ",")
	}

	// Build the command arguments
	args := []string{
		"-w", ff.ffufConfig.Wordlist,
		"-u", normalizedTarget + "/FUZZ",
		"-mc", strings.Join(ff.ffufConfig.StatusCodes, ","),
		"-o", tempFile,
		"-of", "json",
	}

	// Add extensions if specified
	if extensionArg != "" {
		args = append(args, "-e", extensionArg)
	}

	utils.InfoLogger.Printf("Running: ffuf %s", strings.Join(args, " "))

	// Create the command with context
	cmd := exec.CommandContext(ctx, "ffuf", args...)

	// Capture stderr for diagnostics
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// Run the command
	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	// Check if context deadline exceeded (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("ffuf timed out after %s", ff.config.Timeout)
		resultCount := ff.parseAndAppendResults(tempFile)
		if resultCount >= 0 {
			utils.InfoLogger.Printf("Partial results appended: %d URLs", resultCount)
			return nil
		}
		return fmt.Errorf("ffuf timed out after %s with no results", ff.config.Timeout)
	}

	// Check for command errors
	if err != nil && exitCode != 0 {
		stderr := stderrBuf.String()
		utils.WarningLogger.Printf("ffuf exited with code %d: %v", exitCode, err)
		if stderr != "" {
			utils.WarningLogger.Printf("ffuf stderr: %s", stderr)
		}
		// Try to salvage any results
		resultCount := ff.parseAndAppendResults(tempFile)
		if resultCount > 0 {
			utils.WarningLogger.Printf("ffuf encountered an error but produced %d results", resultCount)
			return nil
		}
		return fmt.Errorf("ffuf failed (exit %d): %v", exitCode, err)
	}

	// Parse and append results (zero results is success)
	resultCount := ff.parseAndAppendResults(tempFile)
	if resultCount == 0 {
		utils.InfoLogger.Printf("ffuf completed: 0 results for %s", normalizedTarget)
	} else {
		utils.SuccessLogger.Printf("ffuf completed: %d results for %s", resultCount, normalizedTarget)
	}

	return nil
}

// writeTargetHeader writes a separator and target info to the output file
func (ff *Ffuf) writeTargetHeader() error {
	f, err := os.OpenFile(ff.ffufConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	header := fmt.Sprintf("\n\n%s\n[FFUF] Target: %s\n%s\n",
		strings.Repeat("=", 80),
		ff.config.Target,
		strings.Repeat("=", 80))

	_, err = f.WriteString(header)
	return err
}

// parseAndAppendResults parses the JSON output and appends human-readable results
// Returns the number of results found, or -1 on parse error
func (ff *Ffuf) parseAndAppendResults(tempFile string) int {
	// Check if temp file exists
	data, err := os.ReadFile(tempFile)
	if err != nil {
		return -1
	}

	// Parse JSON
	var result FfufResult
	if err := json.Unmarshal(data, &result); err != nil {
		// If JSON parse fails, return -1 but don't crash
		return -1
	}

	// Open main output file for appending
	f, err := os.OpenFile(ff.ffufConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return -1
	}
	defer f.Close()

	// Write human-readable results
	if len(result.Results) == 0 {
		f.WriteString("(no results)\n")
		return 0
	}

	for _, r := range result.Results {
		line := fmt.Sprintf("[%d] %s\n", r.StatusCode, r.URL)
		f.WriteString(line)
	}
	f.WriteString("\n")

	return len(result.Results)
}

// GetName returns the module name
func (ff *Ffuf) GetName() string {
	return "FFUF (Fast Fuzzer)"
}

// GetOutputPath returns the absolute path to the output file
func (ff *Ffuf) GetOutputPath() (string, error) {
	return filepath.Abs(ff.ffufConfig.OutputFile)
}
