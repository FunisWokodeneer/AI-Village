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
	model := flag.String("model", "qwen2.5:1.5b", "Ollama model name")
	endpoint := flag.String("endpoint", "http://127.0.0.1:11434/api/chat", "Ollama chat endpoint")
	flag.Parse()

	if strings.TrimSpace(*goal) == "" {
		fmt.Fprintln(os.Stderr, "usage: ai-village.exe -goal \"Build a small reporting tool\"")
		os.Exit(2)
	}

	village := NewVillage(*model, *endpoint)
	results, err := village.Run(context.Background(), *goal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(strings.Join(results, "\n\n"))
}
