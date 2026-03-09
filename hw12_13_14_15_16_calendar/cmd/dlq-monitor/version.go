package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	release   = "Version: 1.0.0"
	buildDate = "Build date: 2026-03-07"
	gitHash   = "DLQ Monitor Service"
)

func printVersion() {
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Release   string
		BuildDate string
		GitHash   string
	}{
		Release:   release,
		BuildDate: buildDate,
		GitHash:   gitHash,
	}); err != nil {
		fmt.Printf("error while decode version info: %v\n", err)
	}
}
