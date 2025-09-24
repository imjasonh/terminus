package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chainguard-dev/clog"
	"github.com/google/uuid"

	"github.com/gliderlabs/ssh"

	"terminus/game"
	"terminus/renderer"
	"terminus/screen"
	"terminus/server"
)

var gameServer *server.GameServer

func main() {
	// Parse command line arguments
	mapFile := "maze.map" // Default map
	if len(os.Args) > 1 {
		mapFile = os.Args[1]
	}

	// Load map from file
	worldMap, err := game.LoadMapFromFile(mapFile)
	if err != nil {
		clog.Fatalf("Failed to load map %s: %v", mapFile, err)
	}

	// Initialize game server with 10 player limit
	gameServer = server.NewGameServer(worldMap, 10)

	// Start the global game update loop
	go globalGameLoop()

	// Setup SSH server
	sshServer := &ssh.Server{
		Addr:    ":2222",
		Handler: handleSSHSession,
	}

	clog.Info("Terminus SSH server starting on port 2222...")
	clog.Info("Connect with: ssh -p 2222 localhost")
	clog.Fatalf("ListenAndServe: %v", sshServer.ListenAndServe())
}

// globalGameLoop runs the shared game state updates
func globalGameLoop() {
	ticker := time.NewTicker(time.Second / 30) // 30 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for range ticker.C {
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime).Seconds()
		lastTime = currentTime

		// Update shared game state (projectiles, etc.)
		gameServer.Update(deltaTime)
	}
}

// handleSSHSession handles incoming SSH connections
func handleSSHSession(s ssh.Session) {
	// Generate unique session ID
	sessionID := uuid.New().String()

	// Add player to server
	playerSession, err := gameServer.AddPlayer(sessionID)
	if err != nil {
		fmt.Fprintf(s, "Connection rejected: %s\n", err.Error())
		s.Close()
		return
	}

	// Clean up on disconnect
	defer func() {
		gameServer.RemovePlayer(sessionID)
		clog.Infof("Player %s disconnected", sessionID[:8])
	}()

	clog.Infof("Player %s connected from %s", sessionID[:8], s.RemoteAddr())

	// Get terminal size
	ptyReq, winCh, isPty := s.Pty()
	if !isPty {
		fmt.Fprintf(s, "No PTY requested.\n")
		s.Exit(1)
		return
	}

	// Initialize screen and renderer with PTY dimensions
	width, height := int(ptyReq.Window.Width), int(ptyReq.Window.Height)
	if width <= 0 || height <= 0 {
		width, height = 80, 24 // Default fallback
	}

	gameScreen := screen.NewScreen(width, height)
	gameRenderer := renderer.NewRenderer(width, height)

	// Start player session
	runPlayerSession(s, playerSession, gameScreen, gameRenderer, winCh)
}

// runPlayerSession runs the game loop for a single player
func runPlayerSession(s ssh.Session, playerSession *server.PlayerSession, gameScreen *screen.Screen, gameRenderer *renderer.Renderer, winCh <-chan ssh.Window) {
	player := playerSession.Player

	// Hide cursor and clear screen
	fmt.Fprint(s, "\x1b[?25l\x1b[2J\x1b[H")
	defer fmt.Fprint(s, "\x1b[?25h") // Show cursor on exit

	// Input channel for non-blocking input
	inputCh := make(chan byte, 10)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := s.Read(buf)
			if err != nil {
				if err != io.EOF {
					clog.Infof("Input error for player %s: %v", playerSession.ID[:8], err)
				}
				return
			}
			if n > 0 {
				select {
				case inputCh <- buf[0]:
				default:
					// Drop input if channel is full
				}
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
			if !processPlayerInput(inputCh, player, deltaTime, gameServer, s) {
				return // Player requested exit
			}

			// Create debug message including server info
			playerCount := gameServer.GetPlayerCount()
			activeCount := 0
			var nearestFireball *game.Projectile
			for _, p := range gameServer.ProjectileManager.Projectiles {
				if p.Active && p.Type == game.Fireball {
					activeCount++
					if nearestFireball == nil {
						nearestFireball = p
					}
				}
			}

			debugMsg := fmt.Sprintf("Player: (%.1f,%.1f) | Players: %d/10 | FB: %d",
				player.Position.X, player.Position.Y, playerCount, activeCount)

			if nearestFireball != nil {
				debugMsg = fmt.Sprintf("Player: (%.1f,%.1f) | Players: %d/10 | FB: %d at (%.1f,%.1f)",
					player.Position.X, player.Position.Y, playerCount, activeCount,
					nearestFireball.Position.X, nearestFireball.Position.Y)
			}

			gameScreen.SetDebugMessage(debugMsg)

			// Render the game with shared projectiles
			lights := gameServer.ProjectileManager.GetActiveLights()
			gameRenderer.Render(player, gameServer.Map, gameScreen, lights, gameServer.ProjectileManager.Projectiles)
			fmt.Fprint(s, gameScreen.Render())

		case win := <-winCh:
			// Handle terminal resize
			width, height := int(win.Width), int(win.Height)
			if width > 0 && height > 0 {
				gameScreen = screen.NewScreen(width, height)
				gameRenderer = renderer.NewRenderer(width, height)
			}
		}
	}
}

// processPlayerInput handles input for a single player
func processPlayerInput(inputCh chan byte, player *game.Player, deltaTime float64, gameServer *server.GameServer, s ssh.Session) bool {
	// Process all available input
	for {
		select {
		case key := <-inputCh:
			switch key {
			case 'w', 'W':
				player.MoveForward(deltaTime, gameServer.Map)
			case 's', 'S':
				player.MoveBackward(deltaTime, gameServer.Map)
			case 'a', 'A':
				player.StrafeLeft(deltaTime, gameServer.Map)
			case 'd', 'D':
				player.StrafeRight(deltaTime, gameServer.Map)
			case 'q', 'Q':
				player.RotateRight(deltaTime)
			case 'e', 'E':
				player.RotateLeft(deltaTime)
			case ' ':
				// Shoot fireball (shared projectile system)
				fireball := game.NewFireball(player.Position, player.Direction)
				gameServer.ProjectileManager.AddProjectile(fireball)
			case 27: // ESC key
				fmt.Fprint(s, "\x1b[?25h\x1b[2J\x1b[H") // Show cursor and clear screen
				return false
			case 3: // Ctrl+C
				fmt.Fprint(s, "\x1b[?25h\x1b[2J\x1b[H") // Show cursor and clear screen
				return false
			}
		default:
			return true // No more input to process, continue game loop
		}
	}
}
