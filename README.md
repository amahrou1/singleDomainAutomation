# Single Subdomain Reconnaissance Tool

A powerful, modular Go-based reconnaissance tool focused on comprehensive scanning of single subdomains or IPs for bug hunting and security testing.

## 🚀 Version 2.1 - Latest Features

- ✅ **4 Powerful Modules** - Feroxbuster, Dirsearch, FFUF, and Web Crawling
- ✅ **Web Crawling Module** - Katana, Hakrawler, and GAU with live URL filtering
- ✅ **FFUF Integration** - Fast web fuzzer with custom extensions
- ✅ **Unified Output** - All results in single `fuzzing-output` directory
- ✅ **YAML Configuration** - Easy-to-edit `config.yaml` for all settings
- ✅ **Multiple Input Modes** - Single target (`-d`) or batch scan from file (`-l`)
- ✅ **Sequential Processing** - Complete scan of each target before moving to next
- ✅ **Smart URL Filtering** - Removes duplicates, similar URLs, and filters for live targets only

## 📋 Features

### Module 1: Feroxbuster (Fast Directory/File Bruteforce)
- High-performance directory and file discovery
- Customizable status code filtering (200, 403, 301, 302, 307)
- Configurable threading (default: 50 threads)
- 5-hour timeout protection
- Output: `fuzzing-output/feroxbuster.txt`

### Module 2: Dirsearch (Advanced Path Fuzzing)
- Comprehensive file extension coverage (74+ extensions)
- Intelligent path fuzzing
- Full URL output support
- Recursive scanning option
- Customizable status codes
- Output: `fuzzing-output/dirsearch.txt`

### Module 3: FFUF (Fast Fuzzer)
- Ultra-fast web fuzzing
- Custom extension support (.md, .db, .txt, .xml, .sql, .7z, .zip, .tar.gz, .env)
- Status code filtering (default: 200)
- JSON output format
- Optional recursion support
- Output: `fuzzing-output/ffuf.txt`

### Module 4: Web Crawling & URL Discovery
Comprehensive crawling using multiple tools with intelligent filtering:

**Crawling Tools:**
- **Katana** (3 modes):
  - Standard: JS crawling, form extraction, automatic form filling
  - Parameters: Focus on URLs with query parameters (for XSS/SQLi testing)
  - Passive: Archive-based URL discovery
- **Hakrawler**: Fast crawling with depth 3
- **GAU**: Archive URLs from Wayback Machine, CommonCrawl, OTX, URLScan

**Processing Pipeline:**
1. Combine all crawler outputs
2. Filter unwanted extensions (images, CSS, fonts, media files)
3. Get unique URLs using `anew`
4. Remove similar URLs using `uro`
5. Filter live URLs only using `httpx`
6. Output: `fuzzing-output/crawling-result.txt` (only live, unique URLs)

## 📁 Project Structure

```
singleDomainAutomation/
├── main.go                          # Main entry point with CLI
├── config.yaml                      # User-editable configuration
├── go.mod                           # Go dependencies
├── config/
│   └── config.go                    # Configuration management
├── modules/
│   ├── content_discovery.go         # Feroxbuster module
│   ├── dirsearch.go                 # Dirsearch module
│   ├── ffuf.go                      # FFUF module
│   └── crawling.go                  # Web crawling module
├── utils/
│   └── logger.go                    # Logging utilities
├── fuzzing-output/                  # Unified output directory
│   ├── feroxbuster.txt             # All Feroxbuster results
│   ├── dirsearch.txt               # All Dirsearch results
│   ├── ffuf.txt                    # All FFUF results
│   └── crawling-result.txt         # All crawling results (live URLs)
└── subdomain-recon                  # Compiled binary
```

## 🔧 Prerequisites

### 1. Go (version 1.19 or higher)
```bash
go version
```

### 2. Feroxbuster
```bash
# Install on Linux
curl -sL https://raw.githubusercontent.com/epi052/feroxbuster/main/install-nix.sh | bash

# Verify installation
feroxbuster --version
```

### 3. Dirsearch
```bash
# Install with pip
pip3 install dirsearch

# Verify installation
dirsearch --version
```

### 4. FFUF
```bash
# Install with Go
go install github.com/ffuf/ffuf/v2@latest

# Verify installation
ffuf -V
```

### 5. Web Crawling Tools
```bash
# Katana
go install github.com/projectdiscovery/katana/cmd/katana@latest

# Hakrawler
go install github.com/hakluke/hakrawler@latest

# GAU (GetAllURLs)
go install github.com/lc/gau/v2/cmd/gau@latest

# Anew
go install github.com/tomnomnom/anew@latest

# Uro
pip install uro

# Httpx
go install github.com/projectdiscovery/httpx/cmd/httpx@latest

# Verify installations
katana -version
hakrawler -h
gau -h
anew -h
uro -h
httpx -version
```

### 6. Wordlists

**For Feroxbuster:** Place your wordlist at `/root/myLists/all.txt` or configure in `config.yaml`

**For FFUF:** Uses `/root/SecLists/Discovery/Web-Content/raft-medium-directories.txt` by default

**Recommended wordlist collection:**
```bash
# Clone SecLists
cd /root
git clone https://github.com/danielmiessler/SecLists.git
```

## 🚀 Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd singleDomainAutomation

# Build the binary
go build -o subdomain-recon

# Move to system PATH (optional)
sudo mv subdomain-recon /usr/local/bin/

# Verify installation
subdomain-recon -v
```

### Option 2: Quick Install Script

```bash
cd singleDomainAutomation
chmod +x install.sh
./install.sh
```

## ⚙️ Configuration

Edit `config.yaml` to customize your scans:

```yaml
# Module Enable/Disable
modules:
  feroxbuster: true
  dirsearch: true
  ffuf: true
  crawling: true

# General Settings
general:
  timeout: "5h"  # Maximum timeout for each module

# Feroxbuster Configuration
feroxbuster:
  wordlist: "/root/myLists/all.txt"
  status_codes: [200, 403, 301, 302, 307]
  threads: 50

# Dirsearch Configuration
dirsearch:
  wordlist: ""  # Leave empty for default
  extensions: [php, asp, aspx, jsp, py, txt, ...]
  status_codes: [200, 403, 301, 302]
  threads: 30

# FFUF Configuration
ffuf:
  wordlist: "/root/SecLists/Discovery/Web-Content/raft-medium-directories.txt"
  status_codes: [200]
  extensions: [md, db, txt, xml, sql, 7z, zip, tar.gz, env]
  threads: 40

# Web Crawling Configuration
crawling:
  output_file: "crawling-result.txt"
  katana:
    depth: 5
    concurrency: 20
    rate_limit: 150
```

## 📖 Usage

### Single Target

```bash
# Scan a single subdomain
subdomain-recon -d https://api.example.com

# Scan an IP with non-standard port
subdomain-recon -d https://192.168.1.100:8443
```

### Multiple Targets (Batch Mode)

```bash
# Scan multiple targets from a file
subdomain-recon -l targets.txt
```

**targets.txt format:**
```
https://api.example.com
https://admin.example.com
https://192.168.1.100:8443
http://10.0.0.1:8080
# Comments are supported
```

### Custom Configuration

```bash
# Use custom config file
subdomain-recon -d https://target.com -c custom-config.yaml
```

### Help and Version

```bash
# Show help
subdomain-recon -h

# Show version
subdomain-recon -v
```

## 📊 Output

All results are saved to the `./fuzzing-output/` directory:

```
fuzzing-output/
├── feroxbuster.txt       # All discovered directories/files from all targets
├── dirsearch.txt         # All fuzzing results from all targets
├── ffuf.txt              # All FFUF fuzzing results from all targets
└── crawling-result.txt   # All live URLs from all targets (filtered and unique)
```

Each target's results are clearly separated with headers:

```
================================================================================
[FEROXBUSTER] Target: https://api.example.com
================================================================================
<results here>

================================================================================
[FEROXBUSTER] Target: https://admin.example.com
================================================================================
<results here>
```

## 🔄 Updating the Tool

### Quick Update

```bash
# Navigate to project directory
cd /home/user/singleDomainAutomation

# Pull latest changes
git pull origin claude/subdomain-recon-scanner-XuXXb

# Rebuild
go build -o subdomain-recon

# Install to system PATH
sudo mv subdomain-recon /usr/local/bin/

# Verify update
subdomain-recon -v
```

### Using Update Script

```bash
# Create update script
cat > update.sh << 'EOF'
#!/bin/bash
cd /home/user/singleDomainAutomation
git pull origin claude/subdomain-recon-scanner-XuXXb
go build -o subdomain-recon
sudo mv subdomain-recon /usr/local/bin/
echo "✅ Update complete!"
subdomain-recon -v
EOF

chmod +x update.sh

# Run update
./update.sh
```

## 🎯 Workflow Example

Sequential processing ensures thorough scanning of each target:

```
Target 1: https://api.example.com
  ├─ [1/4] Feroxbuster → fuzzing-output/feroxbuster.txt
  ├─ [2/4] Dirsearch   → fuzzing-output/dirsearch.txt
  ├─ [3/4] FFUF        → fuzzing-output/ffuf.txt
  └─ [4/4] Crawling    → fuzzing-output/crawling-result.txt
      ├─ Katana (standard, params, passive)
      ├─ Hakrawler
      ├─ GAU
      ├─ Filter extensions
      ├─ Anew (unique)
      ├─ Uro (remove similar)
      └─ Httpx (live URLs only)

Target 2: https://admin.example.com
  ├─ [1/4] Feroxbuster (append results)
  ├─ [2/4] Dirsearch (append results)
  ├─ [3/4] FFUF (append results)
  └─ [4/4] Crawling (append results)
```

## 🛡️ Enable/Disable Modules

You can enable or disable any module in `config.yaml`:

```yaml
modules:
  feroxbuster: true   # Set to false to skip
  dirsearch: true
  ffuf: false         # Disabled
  crawling: true
```

## 🔍 Advanced Tips

### 1. Custom Wordlists
```yaml
feroxbuster:
  wordlist: "/path/to/custom/wordlist.txt"
```

### 2. Adjust Timeout
```yaml
general:
  timeout: "3h"  # Reduce for faster scans
```

### 3. Thread Optimization
```yaml
feroxbuster:
  threads: 100  # Increase for faster scanning (requires more resources)
```

### 4. Status Code Filtering
```yaml
feroxbuster:
  status_codes: [200, 201, 301, 302, 401, 403]  # Add more codes
```

## 🐛 Troubleshooting

### Tools Not Found
```bash
# Check if tools are in PATH
which feroxbuster dirsearch ffuf katana hakrawler gau anew uro httpx

# Add Go bin to PATH if needed
export PATH=$PATH:$HOME/go/bin
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
```

### Wordlist Not Found
```bash
# Verify wordlist path
ls -la /root/myLists/all.txt
ls -la /root/SecLists/Discovery/Web-Content/raft-medium-directories.txt

# Update config.yaml with correct path
```

### Permission Denied
```bash
# Give execution permission to binary
chmod +x subdomain-recon

# Or run with sudo for system installation
sudo mv subdomain-recon /usr/local/bin/
```

### Module Timeout
```bash
# Increase timeout in config.yaml
general:
  timeout: "10h"  # Increase for large targets
```

## 📝 Example Output

### Feroxbuster Output
```
200      GET       15l       38w      615c https://api.example.com/admin
403      GET        7l       10w      162c https://api.example.com/config
301      GET        9l       28w      312c https://api.example.com/api
```

### Crawling Output (Live URLs Only)
```
https://api.example.com/v1/users
https://api.example.com/v1/auth/login
https://api.example.com/v2/endpoints
https://api.example.com/docs/api.json
```

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Report bugs
- Suggest new features
- Submit pull requests

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- [Feroxbuster](https://github.com/epi052/feroxbuster)
- [Dirsearch](https://github.com/maurosoria/dirsearch)
- [FFUF](https://github.com/ffuf/ffuf)
- [Katana](https://github.com/projectdiscovery/katana)
- [Hakrawler](https://github.com/hakluke/hakrawler)
- [GAU](https://github.com/lc/gau)
- [Anew](https://github.com/tomnomnom/anew)
- [Uro](https://github.com/s0md3v/uro)
- [Httpx](https://github.com/projectdiscovery/httpx)
- [SecLists](https://github.com/danielmiessler/SecLists)

## 📞 Support

For issues, questions, or feature requests, please open an issue on the project repository.

---

**Happy Bug Hunting! 🐛🔍**
