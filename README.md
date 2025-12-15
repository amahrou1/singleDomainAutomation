# Single Subdomain Reconnaissance Tool

A modular Go-based reconnaissance tool focused on scanning a single subdomain for bug hunting and security testing.

## Version 2.0 - What's New! 🚀

- ✅ **YAML Configuration File** - Easy-to-edit `config.yaml` for all settings
- ✅ **Dirsearch Module** - Advanced path fuzzing with 74+ file extensions
- ✅ **Dual Content Discovery** - Run both Feroxbuster AND Dirsearch
- ✅ **Enhanced CLI** - Help flags, version info, custom config paths
- ✅ **Better Logging** - Module execution summary and success/failure tracking

## Features

### Current Modules

#### Module 1: Feroxbuster (Fast Directory/File Bruteforce)
- High-performance directory and file discovery
- Customizable status code filtering
- Configurable threading
- 5-hour timeout protection
- Output: `feroxbuster.txt`

#### Module 2: Dirsearch (Advanced Path Fuzzing)
- Comprehensive file extension coverage (74+ extensions)
- Intelligent path fuzzing
- Full URL output support
- Recursive scanning option
- Output: `dirsearch.txt`

## Project Structure

```
singleDomainAutomation/
├── main.go                          # Main entry point with CLI
├── config.yaml                      # User-editable configuration
├── go.mod                           # Go dependencies
├── config/
│   └── config.go                    # Configuration management
├── modules/
│   ├── content_discovery.go         # Feroxbuster module
│   └── dirsearch.go                 # Dirsearch module
├── utils/
│   └── logger.go                    # Logging utilities
├── output/                          # Output directory for results
│   ├── feroxbuster.txt             # Feroxbuster results
│   └── dirsearch.txt               # Dirsearch results
└── subdomain-recon                  # Compiled binary
```

## Prerequisites

### 1. Go (version 1.19 or higher)
```bash
go version
```

### 2. Feroxbuster
Fast content discovery tool written in Rust.

```bash
# Install on Linux
curl -sL https://raw.githubusercontent.com/epi052/feroxbuster/main/install-nix.sh | bash

# Or download from releases
wget https://github.com/epi052/feroxbuster/releases/latest/download/x86_64-linux-feroxbuster.tar.gz
tar -xzf x86_64-linux-feroxbuster.tar.gz
sudo mv feroxbuster /usr/local/bin/
chmod +x /usr/local/bin/feroxbuster

# Verify installation
feroxbuster --version
```

### 3. Dirsearch
Advanced web path scanner written in Python.

```bash
# Install with pip
pip3 install dirsearch

# Or clone from GitHub
git clone https://github.com/maurosoria/dirsearch.git
cd dirsearch
pip3 install -r requirements.txt
sudo ln -s $(pwd)/dirsearch.py /usr/local/bin/dirsearch

# Verify installation
dirsearch --version
```

### 4. Wordlist
Place your wordlist at `/root/myLists/all.txt` or configure a different path in `config.yaml`.

**Recommended wordlists:**
- [SecLists](https://github.com/danielmiessler/SecLists)
- [Assetnote Wordlists](https://wordlists.assetnote.io/)
- [FuzzDB](https://github.com/fuzzdb-project/fuzzdb)

## Installation

1. Clone or download this repository
   ```bash
   git clone <repository-url>
   cd singleDomainAutomation
   ```

2. Build the application
   ```bash
   go build -o subdomain-recon main.go
   ```

3. Make it executable (if needed)
   ```bash
   chmod +x subdomain-recon
   ```

## Configuration

All settings are managed through `config.yaml`. Edit this file to customize:

### General Settings
```yaml
general:
  output_dir: "./output"    # Where to save results
  timeout: "5h"             # Max timeout per module (5h, 30m, 2h30m)
```

### Enable/Disable Modules
```yaml
modules:
  feroxbuster: true         # Enable feroxbuster
  dirsearch: true           # Enable dirsearch
```

### Feroxbuster Configuration
```yaml
feroxbuster:
  wordlist: "/root/myLists/all.txt"
  status_codes:             # HTTP codes to include
    - 200
    - 403
    - 301
    - 302
    - 307
  output_file: "feroxbuster.txt"
  threads: 50               # Concurrent threads
  recursion_depth: 0        # 0 = no recursion
  skip_ssl: true            # Skip SSL verification
  no_state: true            # Don't save/load state
```

### Dirsearch Configuration
```yaml
dirsearch:
  wordlist: ""              # Leave empty for default
  extensions:               # 74 file extensions configured
    - php
    - asp
    - aspx
    - jsp
    # ... (see config.yaml for full list)
  status_codes:
    - 200
    - 403
    - 301
    - 302
  output_file: "dirsearch.txt"
  full_url: true            # Show full URLs
  recursive: true           # Follow redirects
  threads: 30               # Concurrent threads
```

## Usage

### Basic Usage
```bash
./subdomain-recon https://subdomain.test.com
```

### With Custom Config
```bash
./subdomain-recon -c custom-config.yaml https://subdomain.test.com
```

### Show Help
```bash
./subdomain-recon -h
```

### Show Version
```bash
./subdomain-recon -v
```

### Run Without Building
```bash
go run main.go https://subdomain.test.com
```

## Command Details

### Feroxbuster Command
The tool executes:
```bash
feroxbuster -u https://subdomain.test.com \
  -w /root/myLists/all.txt \
  -s 200,403,301,302,307 \
  -t 50 \
  -k \
  -n \
  --no-state \
  -o ./output/feroxbuster.txt
```

**Parameters:**
- `-u` : Target URL
- `-w` : Wordlist path
- `-s` : Status codes to include
- `-t` : Number of threads
- `-k` : Skip SSL certificate verification
- `-n` : Don't scan recursively
- `--no-state` : Don't save or load state
- `-o` : Output file path

### Dirsearch Command
The tool executes:
```bash
dirsearch -u https://subdomain.test.com \
  -e php,asp,aspx,jsp,py,txt,conf,... (74 extensions) \
  -i 200,403,301,302 \
  -t 30 \
  --full-url \
  -r \
  -o ./output/dirsearch.txt
```

**Parameters:**
- `-u` : Target URL
- `-e` : File extensions to check
- `-i` : Status codes to include
- `-t` : Number of threads
- `--full-url` : Display full URLs in output
- `-r` : Recursive scanning (follow redirects)
- `-o` : Output file path

## Output

Results are saved in the `./output/` directory:
- `feroxbuster.txt` - Feroxbuster discovered paths
- `dirsearch.txt` - Dirsearch discovered paths

Both files contain:
- Discovered URLs/paths
- HTTP status codes
- Response sizes
- Timestamps

## Error Handling

The tool includes robust error handling:

1. **Missing Tools**: Modules skip gracefully if tools aren't installed
2. **Missing Wordlist**: Module skips with clear error message
3. **Timeout Protection**: 5-hour timeout per module (configurable)
4. **Partial Results**: Saves results even on timeout or errors
5. **Continue on Failure**: Failed modules don't stop other modules

## Example Output

```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     Single Subdomain Reconnaissance Tool                 ║
║     Version: 2.0                                          ║
║     Modules: Feroxbuster + Dirsearch                      ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

[INFO] Target: https://subdomain.test.com
[SUCCESS] Configuration loaded from: config.yaml

[Configuration]
─────────────────────────────────────────────────────────
Target:       https://subdomain.test.com
Output Dir:   ./output
Timeout:      5h0m0s (per module)

[Enabled Modules]
  Feroxbuster:  true
  Dirsearch:    true

[Feroxbuster Settings]
  Wordlist:     /root/myLists/all.txt
  Status Codes: [200 403 301 302 307]
  Threads:      50
  Output:       ./output/feroxbuster.txt

[Dirsearch Settings]
  Wordlist:     (default)
  Extensions:   74 configured
  Status Codes: [200 403 301 302]
  Threads:      30
  Output:       ./output/dirsearch.txt
─────────────────────────────────────────────────────────

============================================================
[INFO] Starting Module: Feroxbuster (Directory/File Bruteforce)
============================================================
[INFO] Running: feroxbuster -u https://subdomain.test.com ...
[INFO] Output will be saved to: ./output/feroxbuster.txt
[INFO] Timeout: 5h0m0s

[feroxbuster scan output...]

[SUCCESS] Module 'Feroxbuster' completed successfully
[SUCCESS] Output saved to: ./output/feroxbuster.txt (Size: 15234 bytes)
============================================================

============================================================
[INFO] Starting Module: Dirsearch (Advanced Path Fuzzing)
============================================================
[INFO] Running: dirsearch -u https://subdomain.test.com ...
[INFO] Output will be saved to: ./output/dirsearch.txt
[INFO] Timeout: 5h0m0s

[dirsearch scan output...]

[SUCCESS] Module 'Dirsearch' completed successfully
[SUCCESS] Output saved to: ./output/dirsearch.txt (Size: 28456 bytes)
============================================================

[Module Execution Summary]
─────────────────────────────────────────────────────────
Total Modules:      2
Successful:         2
Failed/Skipped:     0
─────────────────────────────────────────────────────────

╔═══════════════════════════════════════════════════════════╗
║                     Scan Complete                         ║
╚═══════════════════════════════════════════════════════════╝
Total Execution Time: 25m43s

Results Directory: ./output

💡 Tip: Review the output files for discovered paths and files
```

## Troubleshooting

### Tools Not Found
```
[WARNING] Module skipped: feroxbuster is not installed or not in PATH
```
**Solution:** Install feroxbuster and ensure it's in your PATH

### Wordlist Not Found
```
[WARNING] Module skipped: wordlist not found at /root/myLists/all.txt
```
**Solution:** Update the wordlist path in `config.yaml` or create the wordlist at the specified location

### Configuration Error
```
[ERROR] Failed to load configuration: invalid timeout format '5hours'
```
**Solution:** Use correct time format in `config.yaml` (e.g., "5h", "30m", "2h30m")

### Timeout Issues
If scans timeout frequently:
1. Increase timeout in `config.yaml`: `timeout: "10h"`
2. Reduce thread count to avoid rate limiting
3. Use smaller wordlists for initial scans

## Future Modules (Planned)

- Module 3: Port Scanning (nmap/masscan)
- Module 4: Technology Detection (whatweb/wappalyzer)
- Module 5: SSL/TLS Analysis (testssl.sh)
- Module 6: JavaScript Analysis (subjs/linkfinder)
- Module 7: Parameter Discovery (arjun)
- Module 8: Subdomain Takeover Check

## Development

### Building from Source
```bash
go mod tidy
go build -o subdomain-recon main.go
```

### Adding New Modules

1. Create a new file in `modules/` (e.g., `port_scanner.go`)
2. Implement the module struct and `Run()` method following the existing pattern
3. Add module configuration to `config.yaml`
4. Update `config/config.go` with module-specific config struct
5. Register the module in `main.go` `runModules()` function
6. Test independently before integration
7. Update this README

### Module Template
```go
type NewModule struct {
    config *config.Config
}

func NewNewModule(cfg *config.Config) *NewModule {
    return &NewModule{config: cfg}
}

func (m *NewModule) Run() error {
    moduleName := "New Module Name"
    utils.LogModuleStart(moduleName)
    startTime := time.Now()

    // Check prerequisites
    // Run module logic with timeout
    // Handle errors gracefully

    duration := time.Since(startTime)
    utils.LogModuleEnd(moduleName, err, duration)
    return err
}
```

## Performance Tips

1. **Adjust Thread Count**: Higher threads = faster scans but more resource usage
2. **Use Focused Wordlists**: Smaller, targeted wordlists for specific technologies
3. **Disable Unused Modules**: Set `feroxbuster: false` or `dirsearch: false` in config
4. **Run Modules Separately**: For very large targets, run one module at a time
5. **Monitor Resource Usage**: Watch CPU/memory during scans

## Security & Legal

⚠️ **IMPORTANT**: This tool is for authorized security testing only.

- Always get explicit permission before scanning any target
- Respect rate limits and robots.txt
- Be aware of your country's computer misuse laws
- Use responsibly in bug bounty programs
- Don't use for malicious purposes

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This tool is for authorized security testing only.

## Credits

- **Feroxbuster**: [epi052/feroxbuster](https://github.com/epi052/feroxbuster)
- **Dirsearch**: [maurosoria/dirsearch](https://github.com/maurosoria/dirsearch)

## Support

For issues, questions, or feature requests:
- Open an issue on GitHub
- Check existing issues for solutions
- Provide detailed information (OS, Go version, error messages)

---

**Version**: 2.0
**Last Updated**: 2025-12-15
**Status**: Active Development
