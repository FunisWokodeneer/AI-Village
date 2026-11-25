package main

import (
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("AGS AI Village")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewVillageGame()
	storePath, err := DefaultOrganizationPath()
	if err == nil {
		game.organization, err = OpenOrganizationStore(storePath)
	}
	if err != nil {
		log.Printf("Organization storage unavailable: %v", err)
		game.notice = "Organization storage unavailable; workspace tools are disabled"
		game.noticeUntil = time.Now().Add(12 * time.Second)
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Printf("AI Village stopped: %v", err)
		os.Exit(1)
	}
}
