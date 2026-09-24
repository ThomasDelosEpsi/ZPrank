package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	winW = 520
	winH = 300

	btnW = 140
	btnH = 44
)

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatalf("SDL init error: %v", err)
	}
	defer sdl.Quit()

	// Fenêtre
	window, err := sdl.CreateWindow(
		"Attrape le bouton ! (Go + SDL2)",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		winW, winH,
		sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatalf("CreateWindow: %v", err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		log.Fatalf("CreateRenderer: %v", err)
	}
	defer renderer.Destroy()

	// Bouton: position initiale au centre
	btn := sdl.Rect{X: (winW - btnW) / 2, Y: (winH - btnH) / 2, W: btnW, H: btnH}

	// Affichage: simple rectangle + libellé approximatif
	draw := func() {
		// fond
		renderer.SetDrawColor(240, 240, 240, 255)
		renderer.Clear()

		// bouton
		renderer.SetDrawColor(76, 175, 80, 255) // vert
		renderer.FillRect(&btn)

		// bord bouton
		renderer.SetDrawColor(50, 120, 55, 255)
		renderer.DrawRect(&btn)

		// (Pas de texte raster ici pour rester sans dépendance TTF)
		// On dessine une petite barre blanche au centre pour "simuler" un label
		renderer.SetDrawColor(255, 255, 255, 255)
		label := sdl.Rect{
			X: btn.X + 16, Y: btn.Y + (btn.H/2 - 6),
			W: btn.W - 32, H: 12,
		}
		renderer.FillRect(&label)

		renderer.Present()
	}

	// Utilitaires écran(s)
	getDisplays := func() ([]sdl.Rect, error) {
		n, err := sdl.GetNumVideoDisplays()
		if err != nil {
			return nil, err
		}
		var bounds []sdl.Rect
		for i := 0; i < n; i++ {
			var r sdl.Rect
			if err := sdl.GetDisplayBounds(i, &r); err != nil {
				return nil, err
			}
			bounds = append(bounds, r)
		}
		return bounds, nil
	}

	centerOnDisplay := func(displayIdx int) {
		boundsList, err := getDisplays()
		if err != nil || displayIdx < 0 || displayIdx >= len(boundsList) {
			// Fallback: centrer normal
			window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
			return
		}
		b := boundsList[displayIdx]
		// léger décalage aléatoire pour varier
		offX := rand.Intn(121) - 60 // [-60, +60]
		offY := rand.Intn(81) - 40  // [-40, +40]
		newX := int32(b.X + (b.W-winW)/2 + int32(offX))
		newY := int32(b.Y + (b.H-winH)/2 + int32(offY))
		window.SetPosition(newX, newY)
	}

	currentDisplay := func() int {
		// Déterminer l'écran "actuel" à partir de la position de la fenêtre
		wx, wy := window.GetPosition()
		boundsList, err := getDisplays()
		if err != nil || len(boundsList) == 0 {
			return 0
		}
		for i, r := range boundsList {
			if wx >= r.X && wy >= r.Y && wx < r.X+r.W && wy < r.Y+r.H {
				return i
			}
		}
		return 0
	}

	teleportToNextDisplay := func() {
		boundsList, err := getDisplays()
		if err != nil || len(boundsList) <= 1 {
			return
		}
		idx := currentDisplay()
		next := (idx + 1) % len(boundsList)
		centerOnDisplay(next)
	}

	moveButtonRandomInsideWindow := func() {
		// S'assure que le bouton reste entièrement visible
		w, h := window.GetSize()
		maxX := int(w) - int(btn.W)
		maxY := int(h) - int(btn.H)
		if maxX < 0 {
			maxX = 0
		}
		if maxY < 0 {
			maxY = 0
		}
		var nx, ny int32
		for {
			nx = int32(rand.Intn(maxX + 1))
			ny = int32(rand.Intn(maxY + 1))
			// éviter de retomber quasi au même endroit
			if absI(int(nx)-int(btn.X)) >= 20 || absI(int(ny)-int(btn.Y)) >= 20 {
				break
			}
		}
		btn.X, btn.Y = nx, ny
	}

	// Centrer la fenêtre au démarrage
	centerOnDisplay(currentDisplay())

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				if ev.Type == sdl.KEYDOWN && ev.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}

			case *sdl.MouseMotionEvent:
				// La souris survole le bouton ?
				if pointInRect(int(ev.X), int(ev.Y), btn) {
					// multi-écran ?
					if num, _ := sdl.GetNumVideoDisplays(); num > 1 {
						teleportToNextDisplay()
					} else {
						moveButtonRandomInsideWindow()
					}
				}

			case *sdl.MouseButtonEvent:
				if ev.Type == sdl.MOUSEBUTTONDOWN && ev.Button == sdl.BUTTON_LEFT {
					if pointInRect(int(ev.X), int(ev.Y), btn) {
						// Message de victoire
						_ = sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_INFORMATION,
							"Gagné", "Bravo, tu as réussi !", window)
					}
				}
			}
		}

		draw()
		sdl.Delay(10)
	}
}

func pointInRect(x int, y int, r sdl.Rect) bool {
	return x >= int(r.X) && y >= int(r.Y) && x < int(r.X+r.W) && y < int(r.Y+r.H)
}

func absI(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
