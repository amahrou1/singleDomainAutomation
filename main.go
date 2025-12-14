package main

import (
	"fmt"
	"os"
	"time"

	"subdomain-recon/config"
	"subdomain-recon/modules"
	"subdomain-recon/utils"
)

func main() {
	// Initialize logger
	utils.InitLogger()

	// Print banner
	printBanner()

	// Get target from command line args
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <target>")
		fmt.Println("Example: go run main.go https://subdomain.test.com")
		os.Exit(1)
	}

	target := os.Args[1]

	// Validate target
	if target == "" {
		utils.ErrorLogger.Fatal("Target cannot be empty")
	}

	utils.InfoLogger.Printf("Target: %s", target)

	// Create configuration
	cfg := config.NewConfig(target)

	// Print configuration
	printConfig(cfg)

	// Start total timer
	totalStart := time.Now()

	// Run modules
	runModules(cfg)

	// Print summary
	totalDuration := time.Since(totalStart)
	printSummary(totalDuration)
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Single Subdomain Reconnaissance Tool                 ║
║     Version: 1.0                                          ║
║     Module: Content Discovery (feroxbuster)               ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printConfig(cfg *config.Config) {
	fmt.Println("\n[Configuration]")
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("Target:       %s\n", cfg.Target)
	fmt.Printf("Output Dir:   %s\n", cfg.OutputDir)
	fmt.Printf("Wordlist:     %s\n", cfg.WordlistPath)
	fmt.Printf("Timeout:      %s\n", cfg.Timeout)
	fmt.Println("─────────────────────────────────────────────────────────\n")
}

func runModules(cfg *config.Config) {
	// Module 1: Content Discovery
	if cfg.EnableModules["content_discovery"] {
		contentDiscovery := modules.NewContentDiscovery(cfg)
		err := contentDiscovery.Run()
		if err != nil {
			utils.WarningLogger.Printf("Content Discovery module completed with errors: %v", err)
			utils.InfoLogger.Println("Continuing with next modules (if any)...")
		} else {
			// Get output path
			outputPath, _ := contentDiscovery.GetOutputPath()
			utils.SuccessLogger.Printf("Content Discovery results available at: %s", outputPath)
		}
	}

	// Future modules will be added here
	// Module 2: Port Scanning
	// Module 3: Technology Detection
	// etc.
}

func printSummary(duration time.Duration) {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Scan Summary                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total Execution Time: %s\n", duration)
	fmt.Println("\nRecon completed! Check the output directory for results.")
}
