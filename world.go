package main

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	screenWidth  = 1100
	screenHeight = 720
)

type Point struct {
	X, Y float64
}

type Resident struct {
	Name, Role, Home, Workplace string
	Personality, CurrentAction  string
	DecisionSource              string
	Position, Target            Point
	Color                       color.RGBA
	Speed                       float64
	Thinking                    bool
	LastThought                 time.Time
}

type AgentDecision struct {
	ResidentIndex int
	Activity      string
	Destination   string
	Source        string
}

type Building struct {
	Name, Kind string
	X, Y, W, H float32
	Fill       color.RGBA
}

type VillageGame struct {
	residents             []*Resident
	buildings             []Building
	organization          *OrganizationStore
	paused                bool
	selected              int
	clock                 float64
	lastTick              time.Time
	face                  *text.GoTextFace
	notice                string
	noticeUntil           time.Time
	organizationSummary   string
	organizationSummaryAt time.Time
	rng                   *rand.Rand
	decisionResults       chan AgentDecision
}

func NewVillageGame() *VillageGame {
	src, _ := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	g := &VillageGame{
		selected:        -1,
		lastTick:        time.Now(),
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
		decisionResults: make(chan AgentDecision, 16),
		buildings: []Building{
			{"AGS Commons", "HQ", 425, 260, 250, 145, color.RGBA{67, 91, 119, 255}},
			{"Research House", "RESEARCH", 95, 115, 190, 120, color.RGBA{89, 126, 151, 255}},
			{"Workshop", "BUILD", 805, 105, 190, 130, color.RGBA{155, 111, 75, 255}},
			{"Creative Studio", "DESIGN", 820, 420, 190, 120, color.RGBA{133, 103, 150, 255}},
			{"Garden Cafe", "CAFE", 105, 440, 190, 115, color.RGBA{91, 139, 104, 255}},
		},
	}
	if src != nil {
		g.face = &text.GoTextFace{Source: src, Size: 14}
	}
	g.residents = []*Resident{
		{Name: "Aster", Role: "Chief of Staff", Home: "AGS Commons", Workplace: "AGS Commons", Personality: "calm, organized, thoughtful", CurrentAction: "Checking the day's priorities", Position: Point{535, 335}, Target: Point{535, 335}, Color: color.RGBA{245, 190, 85, 255}, Speed: 48},
		{Name: "Mira", Role: "Research Director", Home: "Research House", Workplace: "Research House", Personality: "curious, precise, patient", CurrentAction: "Reviewing research notes", Position: Point{280, 210}, Target: Point{280, 210}, Color: color.RGBA{115, 196, 220, 255}, Speed: 43},
		{Name: "Atlas", Role: "Engineering Director", Home: "Workshop", Workplace: "Workshop", Personality: "practical, persistent, analytical", CurrentAction: "Working on a prototype", Position: Point{790, 265}, Target: Point{790, 265}, Color: color.RGBA{224, 133, 104, 255}, Speed: 46},
		{Name: "Nova", Role: "Creative Director", Home: "Creative Studio", Workplace: "Creative Studio", Personality: "imaginative, expressive, observant", CurrentAction: "Sketching a new idea", Position: Point{750, 490}, Target: Point{750, 490}, Color: color.RGBA{183, 145, 220, 255}, Speed: 45},
		{Name: "Pip", Role: "Community & Operations", Home: "Garden Cafe", Workplace: "Garden Cafe", Personality: "friendly, energetic, helpful", CurrentAction: "Taking a coffee break", Position: Point{310, 500}, Target: Point{310, 500}, Color: color.RGBA{129, 190, 128, 255}, Speed: 52},
	}
	for _, resident := range g.residents {
		resident.DecisionSource = "demo simulation"
	}
	if os.Getenv("AI_VILLAGE_API_KEY") != "" || os.Getenv("NIM_API_KEY") != "" {
		g.notice = "Provider configured — responses are unverified; project execution is unavailable"
	} else {
		g.notice = "DEMO MODE — scripted activity only; project execution is unavailable"
	}
	g.noticeUntil = time.Now().Add(12 * time.Second)
	return g
}

func (g *VillageGame) Update() error {
	now := time.Now()
	dt := now.Sub(g.lastTick).Seconds()
	if dt > 0.1 {
		dt = 0.1
	}
	g.lastTick = now
	if !now.Before(g.organizationSummaryAt) {
		g.organizationSummaryAt = now.Add(time.Second)
		if g.organization == nil {
			g.organizationSummary = "ORG STORAGE OFFLINE"
		} else {
			organization := g.organization.Snapshot()
			pending := 0
			for _, approval := range organization.Approvals {
				if approval.Status == ApprovalPending {
					pending++
				}
			}
			g.organizationSummary = fmt.Sprintf("ORG  %d TASKS  %d PENDING  %d EVENTS",
				len(organization.Tasks), pending, len(organization.Activity))
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.selected = -1
	}

	for {
		select {
		case decision := <-g.decisionResults:
			if decision.ResidentIndex >= 0 && decision.ResidentIndex < len(g.residents) {
				a := g.residents[decision.ResidentIndex]
				a.Thinking = false
				if decision.Source != "" {
					a.DecisionSource = decision.Source
				}
				if decision.Activity != "" {
					a.CurrentAction = decision.Activity
				}
				if target, ok := destinationPoint(decision.Destination); ok {
					a.Target = target
				}
			}
		default:
			goto decisionsDrained
		}
	}

decisionsDrained:
	if !g.paused {
		g.clock += dt * 12
		for i, a := range g.residents {
			dx, dy := a.Target.X-a.Position.X, a.Target.Y-a.Position.Y
			dist := math.Hypot(dx, dy)
			if dist > 1 {
				step := a.Speed * dt
				if step > dist {
					step = dist
				}
				a.Position.X += dx / dist * step
				a.Position.Y += dy / dist * step
			} else if g.rng.Float64() < dt*0.12 {
				g.chooseAmbientDestination(a)
			}
			if time.Since(a.LastThought) > 35*time.Second && !a.Thinking {
				a.LastThought = now
				a.Thinking = true
				go g.requestDecision(i, *a)
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.selected = -1
		for i, a := range g.residents {
			if math.Hypot(float64(x)-a.Position.X, float64(y)-a.Position.Y) < 26 {
				g.selected = i
				break
			}
		}
	}
	return nil
}

func (g *VillageGame) chooseAmbientDestination(a *Resident) {
	places := []Point{
		{535, 335}, {390, 350}, {710, 350}, {550, 450},
		{300, 350}, {770, 350}, {550, 205}, {550, 555},
	}
	a.Target = places[g.rng.Intn(len(places))]
	actions := []string{
		"walking to the commons", "meeting a teammate",
		"taking a short break", "heading to the cafe",
		"thinking through a task",
	}
	a.CurrentAction = actions[g.rng.Intn(len(actions))]
	a.DecisionSource = "demo simulation"
}

func (g *VillageGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{105, 153, 112, 255})
	vector.DrawFilledRect(screen, 0, 310, screenWidth, 95, color.RGBA{194, 181, 151, 255}, false)
	vector.DrawFilledRect(screen, 485, 0, 130, screenHeight, color.RGBA{194, 181, 151, 255}, false)
	vector.DrawFilledRect(screen, 425, 260, 250, 145, color.RGBA{205, 190, 157, 255}, false)

	for _, b := range g.buildings {
		g.drawBuilding(screen, b)
	}
	g.drawTrees(screen)
	for _, a := range g.residents {
		g.drawResident(screen, a)
	}
	g.drawHUD(screen)
	if g.selected >= 0 {
		g.drawResidentCard(screen, g.residents[g.selected])
	}
	if g.paused {
		g.drawText(screen, "PAUSED", screenWidth/2-35, 34, 20, color.White)
	}
	if time.Now().Before(g.noticeUntil) {
		g.drawText(screen, g.notice, 18, screenHeight-18, 12, color.RGBA{255, 255, 220, 255})
	}
}

func (g *VillageGame) drawBuilding(dst *ebiten.Image, b Building) {
	vector.DrawFilledRect(dst, b.X+5, b.Y+7, b.W, b.H, color.RGBA{47, 67, 52, 100}, false)
	vector.DrawFilledRect(dst, b.X, b.Y, b.W, b.H, b.Fill, false)
	vector.DrawFilledRect(dst, b.X+8, b.Y+8, b.W-16, b.H-16, color.RGBA{232, 218, 183, 255}, false)
	vector.DrawFilledRect(dst, b.X+18, b.Y+18, b.W-36, 17, b.Fill, false)
	vector.DrawFilledRect(dst, b.X+b.W/2-15, b.Y+b.H-37, 30, 37, color.RGBA{102, 77, 61, 255}, false)
	g.drawText(dst, b.Name, float64(b.X+12), float64(b.Y+b.H+20), 13, color.RGBA{255, 255, 240, 255})
}

func (g *VillageGame) drawTrees(dst *ebiten.Image) {
	trees := []Point{
		{45, 60}, {360, 90}, {720, 70}, {1040, 75},
		{45, 640}, {370, 620}, {720, 650}, {1040, 630},
		{360, 450}, {730, 250}, {350, 250}, {760, 570},
	}
	for _, p := range trees {
		vector.DrawFilledCircle(dst, float32(p.X), float32(p.Y), 19, color.RGBA{56, 105, 67, 255}, true)
		vector.DrawFilledCircle(dst, float32(p.X-5), float32(p.Y-6), 12, color.RGBA{88, 145, 83, 255}, true)
	}
}

func (g *VillageGame) drawResident(dst *ebiten.Image, a *Resident) {
	x, y := float32(a.Position.X), float32(a.Position.Y)
	vector.DrawFilledCircle(dst, x, y+7, 13, color.RGBA{38, 55, 45, 100}, true)
	vector.DrawFilledCircle(dst, x, y, 11, a.Color, true)
	vector.DrawFilledCircle(dst, x, y-12, 8, color.RGBA{239, 204, 165, 255}, true)
	vector.DrawFilledCircle(dst, x-3, y-14, 2, color.RGBA{45, 42, 40, 255}, true)
	vector.DrawFilledCircle(dst, x+3, y-14, 2, color.RGBA{45, 42, 40, 255}, true)
	if g.selectedResident(a) {
		vector.StrokeCircle(dst, x, y, 18, 2, color.RGBA{255, 242, 170, 255}, true)
	}
	g.drawText(dst, a.Name, a.Position.X-16, a.Position.Y+26, 12, color.White)
}

func (g *VillageGame) selectedResident(a *Resident) bool {
	return g.selected >= 0 && g.residents[g.selected] == a
}

func (g *VillageGame) drawHUD(dst *ebiten.Image) {
	vector.DrawFilledRect(dst, 0, 0, screenWidth, 48, color.RGBA{27, 39, 48, 235}, false)
	g.drawText(dst, "AGS  /  AI VILLAGE", 18, 29, 19, color.RGBA{245, 239, 218, 255})
	g.drawText(dst, "A living workspace for your AI team", 245, 28, 13, color.RGBA{185, 204, 197, 255})
	g.drawText(dst, g.organizationSummary, 600, 28, 11, color.RGBA{220, 220, 195, 255})
	g.drawText(dst, fmt.Sprintf("DAY 01   %02d:%02d", int(g.clock/60)%24, int(g.clock)%60), 850, 28, 13, color.White)
	g.drawText(dst, fmt.Sprintf("%d RESIDENTS", len(g.residents)), 1000, 28, 12, color.RGBA{224, 224, 205, 255})
	vector.DrawFilledRect(dst, 0, screenHeight-38, screenWidth, 38, color.RGBA{27, 39, 48, 235}, false)
	g.drawText(dst, "Click a resident to inspect  •  SPACE: pause  •  ESC: close profile", 18, screenHeight-14, 12, color.RGBA{210, 220, 210, 255})
}

func (g *VillageGame) drawResidentCard(dst *ebiten.Image, a *Resident) {
	vector.DrawFilledRect(dst, screenWidth-330, 65, 310, 275, color.RGBA{27, 39, 48, 240}, false)
	g.drawText(dst, a.Name+"  /  "+a.Role, screenWidth-310, 95, 16, a.Color)
	g.drawText(dst, "PERSONALITY", screenWidth-310, 130, 11, color.RGBA{158, 190, 178, 255})
	g.drawText(dst, a.Personality, screenWidth-310, 151, 12, color.White)
	g.drawText(dst, "CURRENT ACTIVITY", screenWidth-310, 190, 11, color.RGBA{158, 190, 178, 255})
	g.drawText(dst, a.CurrentAction, screenWidth-310, 213, 12, color.White)
	g.drawText(dst, "HOME", screenWidth-310, 250, 11, color.RGBA{158, 190, 178, 255})
	g.drawText(dst, a.Home, screenWidth-310, 271, 12, color.White)
	source := a.DecisionSource
	if a.Thinking {
		source = "model response pending"
	}
	g.drawText(dst, "Last activity source: "+source, screenWidth-310, 296, 11, color.RGBA{220, 220, 195, 255})
	g.drawText(dst, "Project execution: unavailable", screenWidth-310, 319, 11, color.RGBA{255, 207, 165, 255})
}

func destinationPoint(destination string) (Point, bool) {
	targets := map[string]Point{
		"AGS Commons":     {550, 335},
		"Research House":  {190, 260},
		"Workshop":        {900, 275},
		"Creative Studio": {915, 555},
		"Garden Cafe":     {200, 570},
	}
	point, ok := targets[destination]
	return point, ok
}

func (g *VillageGame) drawText(dst *ebiten.Image, s string, x, y, size float64, c color.Color) {
	if g.face == nil {
		return
	}
	face := *g.face
	face.Size = size
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, &face, op)
}

func (g *VillageGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
