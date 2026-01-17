package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"harjonan.id/user-service/app/config"
	"harjonan.id/user-service/app/router"
)

func init() {
	branch := getGitBranch()

	envFile := getEnvFileFromBranch(branch)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Failed to load env file (%s): %v", envFile, err)
	}

	config.InitLog()
}

func main() {
	port := os.Getenv("PORT")

	init := config.Init()
	app := router.Init(init)

	app.Run(":" + port)
}

func getGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("Warning: git rev-parse failed, defaulting to 'dev'")
		return "dev"
	}
	branch := strings.TrimSpace(out.String())
	log.Println("ENV run in: ", branch)
	return branch
}

func getEnvFileFromBranch(branch string) string {
	switch {
	case branch == "main-aws-stable":
		log.Print("Env prod")
		return "app/config/.env.prod"
	case strings.HasPrefix(branch, "staging"):
		log.Print("Env Staging")
		return "app/config/.env.staging"
	case strings.HasPrefix(branch, "dev"):
		log.Print("Env Dev")
		return "app/config/.env.dev"
	default:
		log.Print("Env local")
		return "app/config/.env.local"
	}
}
