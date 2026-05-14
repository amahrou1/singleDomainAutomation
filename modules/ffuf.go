package modules

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"subdomain-recon/config"
	"subdomain-recon/utils"
)

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

	// Create temp output file for this target
	tempFile := ff.ffufConfig.OutputFile + ".temp"

	// Build the command arguments
	args := []string{
		"-u", ff.config.Target + "/FUZZ",
		"-w", ff.ffufConfig.Wordlist,
		"-mc", strings.Join(ff.ffufConfig.StatusCodes, ","),
		"-t", fmt.Sprintf("%d", ff.ffufConfig.Threads),
		"-o", tempFile,
		"-of", "json", // JSON output for temp file
		"-s",          // Silent mode
	}

	// Add extensions if specified
	if len(ff.ffufConfig.Extensions) > 0 {
		extensions := "." + strings.Join(ff.ffufConfig.Extensions, ",.")
		args = append(args, "-e", extensions)
	}

	// Add recursion if enabled
	if ff.ffufConfig.Recursion {
		args = append(args, "-recursion")
		if ff.ffufConfig.RecursionDepth > 0 {
			args = append(args, "-recursion-depth", fmt.Sprintf("%d", ff.ffufConfig.RecursionDepth))
		}
	}

	utils.InfoLogger.Printf("Running: ffuf %s", strings.Join(args, " "))

	// Create the command with context
	cmd := exec.CommandContext(ctx, "ffuf", args...)

	// Set output to stdout only
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()

	// Append temp file results to main output file
	defer os.Remove(tempFile) // Clean up temp file

	// Check if context deadline exceeded (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		utils.WarningLogger.Printf("ffuf timed out after %s", ff.config.Timeout)
		// Try to append partial results if they exist
		if ff.appendTempToMain(tempFile) {
			utils.InfoLogger.Printf("Partial results appended to %s", ff.ffufConfig.OutputFile)
			return nil
		}
		return fmt.Errorf("ffuf timed out after %s with no results", ff.config.Timeout)
	}

	// Check for other errors
	if err != nil {
		// Try to append results even if there was an error
		if ff.appendTempToMain(tempFile) {
			utils.WarningLogger.Printf("ffuf encountered an error but produced output: %v", err)
			return nil // Don't fail the module if we have output
		}
		return fmt.Errorf("ffuf failed: %v", err)
	}

	// Append temp file to main output file
	if !ff.appendTempToMain(tempFile) {
		return fmt.Errorf("ffuf completed but no output file was generated")
	}

	// Get file info
	fileInfo, err := os.Stat(ff.ffufConfig.OutputFile)
	if err == nil {
		utils.SuccessLogger.Printf("Output appended to: %s (Total size: %d bytes)", ff.ffufConfig.OutputFile, fileInfo.Size())
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

// appendTempToMain appends the temporary file content to the main output file
func (ff *Ffuf) appendTempToMain(tempFile string) bool {
	// Check if temp file exists and has content
	tempInfo, err := os.Stat(tempFile)
	if err != nil || tempInfo.Size() == 0 {
		return false
	}

	// Read temp file
	tempContent, err := os.ReadFile(tempFile)
	if err != nil {
		return false
	}

	// Append to main file
	f, err := os.OpenFile(ff.ffufConfig.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Write(tempContent)
	if err != nil {
		return false
	}

	// Add a newline separator after the results
	f.WriteString("\n")

	return true
}

// GetName returns the module name
func (ff *Ffuf) GetName() string {
	return "FFUF (Fast Fuzzer)"
}

// GetOutputPath returns the absolute path to the output file
func (ff *Ffuf) GetOutputPath() (string, error) {
	return filepath.Abs(ff.ffufConfig.OutputFile)
}
