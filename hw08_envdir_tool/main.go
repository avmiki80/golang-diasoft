package main

import (
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("go-envdir: ")
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatalf("Usage: %s <env-dir> <command> [args...]", args[0])
	}

	path := args[1]
	cmd := args[2:]

	log.Printf("Reading environment from: %s", path)
	log.Printf("Command to execute: %v", cmd)

	envs, err := ReadDir(path)
	if err != nil {
		log.Fatalf("Error reading environment directory '%s': %v", path, err)
	}

	log.Printf("Loaded %d environment variables", len(envs))

	exitCode := RunCmd(cmd, envs)
	log.Printf("Command exited with code: %d", exitCode)

	os.Exit(exitCode)
}
