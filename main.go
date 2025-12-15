package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"subdomain-recon/config"
	"subdomain-recon/modules"
	"subdomain-recon/utils"
)

const (
	Version    = "2.0"
	ConfigFile = "config.yaml"
)

func main() {
	// Initialize logger
	utils.InitLogger()

	// Define flags
	showHelp := flag.Bool("h", false, "Show help message")
	showVersion := flag.Bool("v", false, "Show version")
	configPath := flag.String("c", ConfigFile, "Path to config file")

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("Single Subdomain Reconnaissance Tool v%s\n", Version)
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Get target from command line args
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("❌ Error: Target URL is required\n")
		printUsage()
		os.Exit(1)
	}

	target := args[0]

	// Validate target
	if target == "" {
		utils.ErrorLogger.Fatal("Target cannot be empty")
	}

	utils.InfoLogger.Printf("Target: %s", target)

	// Load configuration from YAML
	cfg, err := config.LoadConfig(target, *configPath)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to load configuration: %v", err)
		utils.ErrorLogger.Printf("Make sure '%s' exists and is valid", *configPath)
		os.Exit(1)
	}

	utils.SuccessLogger.Printf("Configuration loaded from: %s", *configPath)

	// Print configuration
	printConfig(cfg)

	// Start total timer
	totalStart := time.Now()

	// Run modules
	runModules(cfg)

	// Print summary
	totalDuration := time.Since(totalStart)
	printSummary(totalDuration, cfg)
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Single Subdomain Reconnaissance Tool                 ║
║     Version: ` + Version + `                                          ║
║     Modules: Feroxbuster + Dirsearch                      ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printUsage() {
	fmt.Println("Usage: ./subdomain-recon [options] <target>")
	fmt.Println("\nExample:")
	fmt.Println("  ./subdomain-recon https://subdomain.test.com")
	fmt.Println("  ./subdomain-recon -c custom-config.yaml https://subdomain.test.com")
	fmt.Println("\nOptions:")
	fmt.Println("  -h          Show help message")
	fmt.Println("  -v          Show version")
	fmt.Println("  -c <file>   Path to config file (default: config.yaml)")
}

func printHelp() {
	printBanner()
	printUsage()
	fmt.Println("\nDescription:")
	fmt.Println("  A modular reconnaissance tool for scanning a single subdomain.")
	fmt.Println("  Focuses on content discovery and path fuzzing.")
	fmt.Println("\nModules:")
	fmt.Println("  1. Feroxbuster - Fast directory/file bruteforcing")
	fmt.Println("  2. Dirsearch   - Advanced path fuzzing with multiple extensions")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Edit 'config.yaml' to customize:")
	fmt.Println("    - Wordlists")
	fmt.Println("    - Status codes")
	fmt.Println("    - Extensions (dirsearch)")
	fmt.Println("    - Timeout settings")
	fmt.Println("    - Enable/disable modules")
	fmt.Println("    - Output file names")
	fmt.Println("\nOutput:")
	fmt.Println("  Results are saved to ./output/ directory:")
	fmt.Println("    - feroxbuster.txt - Feroxbuster results")
	fmt.Println("    - dirsearch.txt   - Dirsearch results")
	fmt.Println("\nFor more information, see README.md")
}

func printConfig(cfg *config.Config) {
	fmt.Println("\n[Configuration]")
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("Target:       %s\n", cfg.Target)
	fmt.Printf("Output Dir:   %s\n", cfg.OutputDir)
	fmt.Printf("Timeout:      %s (per module)\n", cfg.Timeout)
	fmt.Println("\n[Enabled Modules]")
	fmt.Printf("  Feroxbuster:  %v\n", cfg.EnableModules["feroxbuster"])
	fmt.Printf("  Dirsearch:    %v\n", cfg.EnableModules["dirsearch"])

	if cfg.EnableModules["feroxbuster"] {
		fxCfg := cfg.NewFeroxbusterConfig()
		fmt.Println("\n[Feroxbuster Settings]")
		fmt.Printf("  Wordlist:     %s\n", fxCfg.Wordlist)
		fmt.Printf("  Status Codes: %s\n", fxCfg.StatusCodes)
		fmt.Printf("  Threads:      %d\n", fxCfg.Threads)
		fmt.Printf("  Output:       %s\n", fxCfg.OutputFile)
	}

	if cfg.EnableModules["dirsearch"] {
		dsCfg := cfg.NewDirsearchConfig()
		fmt.Println("\n[Dirsearch Settings]")
		if dsCfg.Wordlist != "" {
			fmt.Printf("  Wordlist:     %s\n", dsCfg.Wordlist)
		} else {
			fmt.Printf("  Wordlist:     (default)\n")
		}
		fmt.Printf("  Extensions:   %d configured\n", len(dsCfg.Extensions))
		fmt.Printf("  Status Codes: %s\n", dsCfg.StatusCodes)
		fmt.Printf("  Threads:      %d\n", dsCfg.Threads)
		fmt.Printf("  Output:       %s\n", dsCfg.OutputFile)
	}

	fmt.Println("─────────────────────────────────────────────────────────\n")
}

func runModules(cfg *config.Config) {
	moduleCount := 0
	successCount := 0
	failCount := 0

	// Module 1: Feroxbuster
	if cfg.EnableModules["feroxbuster"] {
		moduleCount++
		feroxbuster := modules.NewFeroxbuster(cfg)
		err := feroxbuster.Run()
		if err != nil {
			failCount++
			utils.WarningLogger.Printf("Feroxbuster module completed with errors: %v", err)
			utils.InfoLogger.Println("Continuing with next modules...")
		} else {
			successCount++
			outputPath, _ := feroxbuster.GetOutputPath()
			utils.SuccessLogger.Printf("Feroxbuster results: %s", outputPath)
		}
	}

	// Module 2: Dirsearch
	if cfg.EnableModules["dirsearch"] {
		moduleCount++
		dirsearch := modules.NewDirsearch(cfg)
		err := dirsearch.Run()
		if err != nil {
			failCount++
			utils.WarningLogger.Printf("Dirsearch module completed with errors: %v", err)
			utils.InfoLogger.Println("Continuing with next modules...")
		} else {
			successCount++
			outputPath, _ := dirsearch.GetOutputPath()
			utils.SuccessLogger.Printf("Dirsearch results: %s", outputPath)
		}
	}

	// Future modules will be added here
	// Module 3: Port Scanning
	// Module 4: Technology Detection
	// etc.

	// Print module summary
	fmt.Println("\n[Module Execution Summary]")
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("Total Modules:      %d\n", moduleCount)
	fmt.Printf("Successful:         %d\n", successCount)
	fmt.Printf("Failed/Skipped:     %d\n", failCount)
	fmt.Println("─────────────────────────────────────────────────────────")
}

func printSummary(duration time.Duration, cfg *config.Config) {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Scan Complete                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total Execution Time: %s\n", duration)
	fmt.Printf("\nResults Directory: %s\n", cfg.OutputDir)
	fmt.Println("\n💡 Tip: Review the output files for discovered paths and files")
}
