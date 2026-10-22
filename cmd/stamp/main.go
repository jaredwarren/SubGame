package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "bump":
		if len(os.Args) < 3 {
			fmt.Println("usage: stamp bump [patch|minor|major]")
			os.Exit(1)
		}
		bumpVersion(os.Args[2])
	case "version":
		fmt.Println(getFullVersion())
	case "apply":
		version := ""
		if len(os.Args) >= 3 {
			version = os.Args[2]
		} else {
			version = getFullVersion()
		}
		applyVersion(version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SubGame build stamper")
	fmt.Println("Commands:")
	fmt.Println("  stamp bump [patch|minor|major]  Increment VERSION file")
	fmt.Println("  stamp version                   Print full version (VERSION + git short hash)")
	fmt.Println("  stamp apply [version]           Stamp version into web/sw.js and web/index.html")
}

func getBaseVersion() string {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return "0.1.0"
	}
	ver := strings.TrimSpace(string(data))
	if ver == "" {
		return "0.1.0"
	}
	return ver
}

func getGitHash() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	hash := strings.TrimSpace(string(out))
	if hash == "" {
		return "dev"
	}
	return hash
}

func getFullVersion() string {
	base := getBaseVersion()
	hash := getGitHash()
	return fmt.Sprintf("%s-%s", base, hash)
}

func bumpVersion(part string) {
	current := getBaseVersion()
	parts := strings.Split(current, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch strings.ToLower(part) {
	case "patch":
		patch++
	case "minor":
		minor++
		patch = 0
	case "major":
		major++
		minor = 0
		patch = 0
	default:
		fmt.Printf("Unknown bump part: %s (expected patch, minor, or major)\n", part)
		os.Exit(1)
	}

	newVer := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	if err := os.WriteFile("VERSION", []byte(newVer+"\n"), 0644); err != nil {
		fmt.Printf("Failed to write VERSION: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Bumped version: %s -> %s\n", current, newVer)
}

func applyVersion(version string) {
	// Stamp web/sw.js
	swFile := "web/sw.js"
	swContent, err := os.ReadFile(swFile)
	if err != nil {
		fmt.Printf("Warning: failed to read %s: %v\n", swFile, err)
	} else {
		re := regexp.MustCompile(`const BUILD_VERSION = ".*?";`)
		var updated []byte
		if re.Match(swContent) {
			updated = re.ReplaceAll(swContent, []byte(fmt.Sprintf(`const BUILD_VERSION = "%s";`, version)))
		} else {
			// If not yet present, insert right below header
			updated = []byte(fmt.Sprintf("/* SubGame service worker */\nconst BUILD_VERSION = \"%s\";\n%s", version, string(swContent)))
		}
		if err := os.WriteFile(swFile, updated, 0644); err != nil {
			fmt.Printf("Warning: failed to write %s: %v\n", swFile, err)
		} else {
			fmt.Printf("Stamped %s with version %s\n", swFile, version)
		}
	}

	// Stamp web/index.html
	indexFile := "web/index.html"
	indexContent, err := os.ReadFile(indexFile)
	if err != nil {
		fmt.Printf("Warning: failed to read %s: %v\n", indexFile, err)
	} else {
		contentStr := string(indexContent)

		// Replace const BUILD_VERSION
		reConst := regexp.MustCompile(`const BUILD_VERSION = ".*?";`)
		if reConst.MatchString(contentStr) {
			contentStr = reConst.ReplaceAllString(contentStr, fmt.Sprintf(`const BUILD_VERSION = "%s";`, version))
		}

		// Replace <div id="versionTag">...</div>
		reTag := regexp.MustCompile(`<div id="versionTag">.*?</div>`)
		if reTag.MatchString(contentStr) {
			contentStr = reTag.ReplaceAllString(contentStr, fmt.Sprintf(`<div id="versionTag">v%s</div>`, version))
		}

		if err := os.WriteFile(indexFile, []byte(contentStr), 0644); err != nil {
			fmt.Printf("Warning: failed to write %s: %v\n", indexFile, err)
		} else {
			fmt.Printf("Stamped %s with version %s\n", indexFile, version)
		}
	}
}
