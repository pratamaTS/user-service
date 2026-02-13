package helpers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func detectBranchFromGit() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("git rev-parse failed: %v", err)
		return ""
	}
	branch := strings.TrimSpace(out.String())
	if branch == "HEAD" {
		// detached state
		return ""
	}
	return branch
}

// mapBranchToEnv maps a git branch name to app env keyword.
func mapBranchToEnv(branch string) string {
	switch branch {
	case "dev-aws-stable":
		return "development"
	case "staging-aws-stable":
		return "staging"
	case "main-aws-stable":
		return "production"
	case "":
		return "local"
	default:
		return "local"
	}
}

// loadEnvForEnvName searches typical locations and loads the first .env file found.
func loadEnvForEnvName(envName string) (string, error) {
	candidates := []string{
		filepath.Join("/run/env", ".env."+envName), // best: mounted from host
		filepath.Join("config", ".env."+envName),
		".env." + envName,
		".env.current",
		".env",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err != nil {
				return "", fmt.Errorf("found %s but failed to load: %w", p, err)
			}
			return p, nil
		}
	}
	return "", fmt.Errorf("no .env file found for env=%s", envName)
}

// initEnv figures out which env to use and loads it.
func InitEnv() (string, string) {
	// Allow manual override via APP_ENV if you ever need it.
	envName := strings.TrimSpace(os.Getenv("APP_ENV"))
	if envName == "" {
		branch := detectBranchFromGit()
		envName = mapBranchToEnv(branch)
		log.Printf("Branch detected: %q → APP_ENV=%q", branch, envName)
	} else {
		log.Printf("APP_ENV preset: %q", envName)
	}

	loadedFrom, err := loadEnvForEnvName(envName)
	if err != nil {
		log.Printf("ENV load warning: %v (continuing with process env only)", err)
	} else {
		log.Printf("Loaded env from: %s", loadedFrom)
	}
	// Export APP_ENV for downstream code if useful
	_ = os.Setenv("APP_ENV", envName)
	return envName, loadedFrom
}
