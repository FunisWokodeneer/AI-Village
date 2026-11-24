//go:build windows

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const (
	defaultModel    = "meta/llama-3.1-70b-instruct"
	defaultEndpoint = "https://integrate.api.nvidia.com/v1/chat/completions"
)

func main() {
	defer func() {
		if recovered := recover(); recovered != nil {
			walk.MsgBox(nil, "AI Village could not start", fmt.Sprint(recovered), walk.MsgBoxIconError)
		}
	}()

	if _, err := (MainWindow{
		AssignTo: &mainWindow,
		Title:   "AI Village",
		MinSize: Size{Width: 720, Height: 680},
		Size:    Size{Width: 900, Height: 780},
		Layout:  VBox{Margins: Margins{Left: 16, Top: 16, Right: 16, Bottom: 16}, Spacing: 8},
		Children: []Widget{
			Label{Text: "AI Village", Font: Font{PointSize: 18, Bold: true}},
			Label{Text: "Give your four AI agents a goal. They will research, plan, build, and review it in sequence."},
			Label{Text: "NVIDIA API key (required for each launch):"},
			LineEdit{AssignTo: &apiKeyEdit, PasswordMode: true},
			Label{Text: "Project goal:"},
			TextEdit{AssignTo: &goalEdit, MinSize: Size{Width: 500, Height: 100}},
			Label{Text: "Model:"},
			LineEdit{AssignTo: &modelEdit, Text: defaultModel},
			Label{Text: "API endpoint:"},
			LineEdit{AssignTo: &endpointEdit, Text: defaultEndpoint},
			PushButton{
				AssignTo: &runButton,
				Text:     "Run AI Village",
				OnClicked: runVillage,
			},
			ProgressBar{AssignTo: &progressBar, MinValue: 0, MaxValue: 4, Value: 0},
			Label{AssignTo: &statusLabel, Text: "Ready. Enter your API key and project goal."},
			Label{Text: "Agent results:"},
			TextEdit{AssignTo: &outputEdit, ReadOnly: true, MinSize: Size{Width: 500, Height: 240}},
		},
	}.Run()); err != nil {
		walk.MsgBox(nil, "AI Village could not start", err.Error(), walk.MsgBoxIconError)
	}
}

var (
	apiKeyEdit  *walk.LineEdit
	goalEdit    *walk.TextEdit
	modelEdit   *walk.LineEdit
	endpointEdit *walk.LineEdit
	runButton   *walk.PushButton
	progressBar *walk.ProgressBar
	statusLabel *walk.Label
	outputEdit  *walk.TextEdit
	mainWindow  *walk.MainWindow
)

func runVillage() {
	goal := strings.TrimSpace(goalEdit.Text())
	apiKey := strings.TrimSpace(apiKeyEdit.Text())
	model := strings.TrimSpace(modelEdit.Text())
	endpoint := strings.TrimSpace(endpointEdit.Text())

	if goal == "" {
		walk.MsgBox(mainWindow, "Project goal required", "Please enter a goal for AI Village.", walk.MsgBoxIconWarning)
		return
	}
	if apiKey == "" {
		walk.MsgBox(mainWindow, "API key required", "Enter your NVIDIA API key to run the agents.", walk.MsgBoxIconWarning)
		return
	}

	if model == "" {
		model = defaultModel
	}
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	runButton.SetEnabled(false)
	progressBar.SetValue(0)
	outputEdit.SetText("")
	statusLabel.SetText("Starting the agents...")
	village := NewVillage(model, endpoint, apiKey)

	go func() {
		for i, agent := range Agents {
			step := i + 1
			currentAgent := agent
			status := fmt.Sprintf("Agent %d of 4: %s is working...", step, currentAgent.Name)
			mainWindow.Synchronize(func() {
				statusLabel.SetText(status)
			})

			result, err := village.ask(context.Background(), currentAgent, goal)
			if err != nil {
				message := fmt.Sprintf("%s failed: %v", currentAgent.Name, err)
				mainWindow.Synchronize(func() {
					statusLabel.SetText("Run stopped because an agent returned an error.")
					outputEdit.SetText(outputEdit.Text() + message + "\r\n\r\n")
					runButton.SetEnabled(true)
				})
				return
			}

			section := fmt.Sprintf("## %s — %s\r\n%s\r\n\r\n", currentAgent.Name, currentAgent.Specialty, result)
			mainWindow.Synchronize(func() {
				outputEdit.SetText(outputEdit.Text() + section)
				progressBar.SetValue(step)
				if step == 4 {
					statusLabel.SetText("All four agents finished successfully.")
					runButton.SetEnabled(true)
				}
			})
		}
	}()
}
