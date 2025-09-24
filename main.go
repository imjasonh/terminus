package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/term"

	"terminus/game"
	"terminus/renderer"
	"terminus/screen"
)

func main() {
	// Parse command line arguments
	mapFile := "maze.map" // Default map
	if len(os.Args) > 1 {
		mapFile = os.Args[1]
	}

	// Get terminal size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Default fallback size
		width, height = 80, 24
	}

	// Load map from file
	worldMap, err := game.LoadMapFromFile(mapFile)
	if err != nil {
		log.Fatalf("Failed to load map %s: %v", mapFile, err)
	}

	// Set spawn position based on map
	var spawnX, spawnY float64
	switch mapFile {
	case "cave.map":
		spawnX, spawnY = 12.0, 12.0 // Center of cave
	default:
		spawnX, spawnY = 1.5, 1.5 // Maze entrance
	}

	// Initialize game components
	player := game.NewPlayer(spawnX, spawnY)
	projectileManager := game.NewProjectileManager()
	gameScreen := screen.NewScreen(width, height)
	gameRenderer := renderer.NewRenderer(width, height)

	// Set terminal to raw mode for real-time input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal("Failed to set raw terminal mode:", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Hide cursor and clear screen
	fmt.Print("\x1b[?25l\x1b[2J\x1b[H")
	defer fmt.Print("\x1b[?25h") // Show cursor on exit

	// Input channel for non-blocking input
	inputCh := make(chan byte, 1)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			select {
			case inputCh <- buf[0]:
			default:
				// Drop input if channel is full
			}
		}
	}()

	// Game loop
	ticker := time.NewTicker(time.Second / 30) // 30 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		select {
		case <-ticker.C:
			currentTime := time.Now()
			deltaTime := currentTime.Sub(lastTime).Seconds()
			lastTime = currentTime

			// Process input
			processInput(inputCh, player, deltaTime, worldMap, projectileManager)

			// Update projectiles
			projectileManager.Update(deltaTime, worldMap)

			// Create debug message for HUD
			activeCount := 0
			var nearestFireball *game.Projectile
			for _, p := range projectileManager.Projectiles {
				if p.Active && p.Type == game.Fireball {
					activeCount++
					if nearestFireball == nil {
						nearestFireball = p
					}
				}
			}

			debugMsg := fmt.Sprintf("Player: (%.1f,%.1f) | Fireballs: %d",
				player.Position.X, player.Position.Y, activeCount)

			if nearestFireball != nil {
				debugMsg = fmt.Sprintf("Player: (%.1f,%.1f) | FB: %d at (%.1f,%.1f) Life: %.1f",
					player.Position.X, player.Position.Y, activeCount,
					nearestFireball.Position.X, nearestFireball.Position.Y, nearestFireball.Life)
			}

			gameScreen.SetDebugMessage(debugMsg)

			// Render the game with lighting
			lights := projectileManager.GetActiveLights()
			gameRenderer.Render(player, worldMap, gameScreen, lights, projectileManager.Projectiles)
			fmt.Print(gameScreen.Render())

		}
	}
}

func processInput(inputCh chan byte, player *game.Player, deltaTime float64, worldMap *game.Map, projectileManager *game.ProjectileManager) {
	// Process all available input
	for {
		select {
		case key := <-inputCh:
			switch key {
			case 'w', 'W':
				player.MoveForward(deltaTime, worldMap)
			case 's', 'S':
				player.MoveBackward(deltaTime, worldMap)
			case 'a', 'A':
				player.StrafeLeft(deltaTime, worldMap)
			case 'd', 'D':
				player.StrafeRight(deltaTime, worldMap)
			case 'q', 'Q':
				player.RotateRight(deltaTime)
			case 'e', 'E':
				player.RotateLeft(deltaTime)
			case ' ':
				// Shoot fireball
				fireball := game.NewFireball(player.Position, player.Direction)
				projectileManager.AddProjectile(fireball)
			case 27: // ESC key
				fmt.Print("\x1b[?25h\x1b[2J\x1b[H") // Show cursor and clear screen
				os.Exit(0)
			case 3: // Ctrl+C
				fmt.Print("\x1b[?25h\x1b[2J\x1b[H") // Show cursor and clear screen
				os.Exit(0)
			}
		default:
			return // No more input to process
		}
	}
}
