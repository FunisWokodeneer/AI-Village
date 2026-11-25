package main

import "testing"

func TestNewVillageGameCreatesInitialResidents(t *testing.T) {
	game := NewVillageGame()
	if got, want := len(game.residents), 5; got != want {
		t.Fatalf("resident count = %d, want %d", got, want)
	}
	if game.residents[0].Name != "Aster" {
		t.Fatalf("first resident = %q, want Aster", game.residents[0].Name)
	}
	if game.residents[0].DecisionSource != "demo simulation" {
		t.Fatalf("initial decision source = %q, want simulated activity", game.residents[0].DecisionSource)
	}
	game.chooseAmbientDestination(game.residents[0])
	if game.residents[0].DecisionSource != "demo simulation" {
		t.Fatalf("ambient activity source = %q, want simulated activity", game.residents[0].DecisionSource)
	}
}

func TestDestinationPointOnlyAcceptsKnownBuildings(t *testing.T) {
	for _, name := range []string{
		"AGS Commons",
		"Research House",
		"Workshop",
		"Creative Studio",
		"Garden Cafe",
	} {
		if _, ok := destinationPoint(name); !ok {
			t.Errorf("destinationPoint(%q) rejected a known building", name)
		}
	}
	if _, ok := destinationPoint("Unknown Place"); ok {
		t.Fatal("destinationPoint accepted an unknown building")
	}
}
