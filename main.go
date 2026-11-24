package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("AGS AI Village")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewVillageGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Printf("AI Village stopped: %v", err)
		os.Exit(1)
	}
}
