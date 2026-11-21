package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Agent struct {
	Name      string
	Specialty string
	Prompt    string
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Message OllamaMessage `json:"message"`
	Error   string        `json:"error"`
}

var Agents = []Agent{
	{Name: "Researcher", Specialty: "facts and constraints", Prompt: "Only research facts, requirements, constraints, and risks. Do not plan or implement."},
	{Name: "Planner", Specialty: "project planning", Prompt: "Only break the goal into ordered, achievable steps. Do not research, write code, or review."},
	{Name: "Builder", Specialty: "implementation", Prompt: "Only describe concrete implementation details and code. Do not research, plan, or review."},
	{Name: "Reviewer", Specialty: "quality review", Prompt: "Only review the proposed result for correctness, risks, and missing checks. Do not implement or plan."},
}

type Village struct {
	Model    string
	Endpoint string
	Client   *http.Client
}

func (v Village) Run(ctx context.Context, goal string) ([]string, error) {
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return nil, fmt.Errorf("goal is required")
	}

	results := make([]string, 0, len(Agents))
	for _, agent := range Agents {
		result, err := v.ask(ctx, agent, goal)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %w", agent.Name, err)
		}
		results = append(results, fmt.Sprintf("## %s — %s\n%s", agent.Name, agent.Specialty, result))
	}
	return results, nil
}

func (v Village) ask(ctx context.Context, agent Agent, goal string) (string, error) {
	system := agent.Prompt + " Stay within your specialty and return concise output."
	body := ollamaRequest{
		Model: v.Model,
		Messages: []OllamaMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: goal},
		},
		Stream: false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if result.Error != "" {
			return "", fmt.Errorf("AI service returned %s: %s", resp.Status, result.Error)
		}
		return "", fmt.Errorf("AI service returned %s", resp.Status)
	}
	if strings.TrimSpace(result.Message.Content) == "" {
		return "", fmt.Errorf("AI service returned an empty response")
	}
	return strings.TrimSpace(result.Message.Content), nil
}

func NewVillage(model, endpoint string) Village {
	if strings.TrimSpace(model) == "" {
		model = "qwen2.5:1.5b"
	}
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "http://127.0.0.1:11434/api/chat"
	}
	return Village{
		Model:    model,
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}
