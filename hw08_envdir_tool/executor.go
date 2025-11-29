package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const successCode = 0

const errorCode = 1

func runCommand(command string, args []string) (int, error) {
	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return errorCode, fmt.Errorf("failed to execute command: %w", err)
	}

	return successCode, nil
}

func processEnv(env Environment) error {
	var err error
	var processingErrors []error
	for name, envValue := range env {
		if envValue.NeedRemove {
			err = os.Unsetenv(name)
		} else {
			err = os.Setenv(name, envValue.Value)
		}
		if err != nil {
			processingErrors = append(processingErrors, err)
		}
	}
	if len(processingErrors) > 0 {
		err = fmt.Errorf("errors processing some environments variables: %v", processingErrors)
	}
	return err
}

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		log.Printf("ERROR: No command specified")
		return errorCode
	}

	err := processEnv(env)
	if err != nil {
		log.Printf("ERROR: Failed to process environment variables: %v", err)
		return errorCode
	}

	exitCode, err := runCommand(cmd[0], cmd[1:])
	if err != nil {
		log.Printf("ERROR: Command execution failed: %v", err)
		return exitCode
	}

	return exitCode
}
