package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	goal := flag.String("goal", "", "goal for the AI Village")
	model := flag.String("model", "meta/llama-3.1-70b-instruct", "NVIDIA NIM model name")
	endpoint := flag.String("endpoint", "https://integrate.api.nvidia.com/v1/chat/completions", "NVIDIA NIM chat endpoint")
	apiKeyFlag := flag.String("api-key", "", "NVIDIA NIM API key (prefer the NIM_API_KEY environment variable)")
	flag.Parse()

	if strings.TrimSpace(*goal) == "" {
		fmt.Fprintln(os.Stderr, "usage: ai-village.exe -goal \"Build a small reporting tool\"")
		os.Exit(2)
	}

	apiKey := strings.TrimSpace(*apiKeyFlag)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("NIM_API_KEY"))
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "missing NVIDIA NIM API key: set NIM_API_KEY or use -api-key")
		os.Exit(2)
	}

	village := NewVillage(*model, *endpoint, apiKey)
	results, err := village.Run(context.Background(), *goal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.Join(results, "\n\n"))
}
