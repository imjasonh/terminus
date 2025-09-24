package server

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"terminus/game"
)

// GameServer holds the shared state for all connected players
type GameServer struct {
	Map               *game.Map
	ProjectileManager *game.ProjectileManager
	Players           map[string]*PlayerSession
	PlayersMutex      sync.RWMutex
	MaxPlayers        int
}

// PlayerSession represents a connected player's session
type PlayerSession struct {
	ID          string
	Player      *game.Player
	Connected   bool
	ConnectedAt time.Time
}

// NewGameServer creates a new game server instance
func NewGameServer(worldMap *game.Map, maxPlayers int) *GameServer {
	return &GameServer{
		Map:               worldMap,
		ProjectileManager: game.NewProjectileManager(),
		Players:           make(map[string]*PlayerSession),
		MaxPlayers:        maxPlayers,
	}
}

// AddPlayer adds a new player to the server
func (gs *GameServer) AddPlayer(sessionID string) (*PlayerSession, error) {
	gs.PlayersMutex.Lock()
	defer gs.PlayersMutex.Unlock()

	// Check player limit
	if len(gs.Players) >= gs.MaxPlayers {
		return nil, fmt.Errorf("server full: max %d players", gs.MaxPlayers)
	}

	// Find random spawn point
	spawnX, spawnY := gs.findRandomSpawnPoint()

	// Create new player
	player := game.NewPlayer(spawnX, spawnY)
	session := &PlayerSession{
		ID:          sessionID,
		Player:      player,
		Connected:   true,
		ConnectedAt: time.Now(),
	}

	gs.Players[sessionID] = session
	return session, nil
}

// RemovePlayer removes a player from the server
func (gs *GameServer) RemovePlayer(sessionID string) {
	gs.PlayersMutex.Lock()
	defer gs.PlayersMutex.Unlock()

	if session, exists := gs.Players[sessionID]; exists {
		session.Connected = false
		delete(gs.Players, sessionID)
	}
}

// GetPlayerCount returns the current number of connected players
func (gs *GameServer) GetPlayerCount() int {
	gs.PlayersMutex.RLock()
	defer gs.PlayersMutex.RUnlock()
	return len(gs.Players)
}

// GetPlayerSession returns a player session by ID
func (gs *GameServer) GetPlayerSession(sessionID string) (*PlayerSession, bool) {
	gs.PlayersMutex.RLock()
	defer gs.PlayersMutex.RUnlock()
	session, exists := gs.Players[sessionID]
	return session, exists
}

// findRandomSpawnPoint finds a random empty location on the map
func (gs *GameServer) findRandomSpawnPoint() (float64, float64) {
	// Find all empty spaces (value 0)
	var emptySpaces [][2]int

	for y := 0; y < len(gs.Map.Grid); y++ {
		for x := 0; x < len(gs.Map.Grid[y]); x++ {
			if gs.Map.Grid[y][x] == 0 {
				emptySpaces = append(emptySpaces, [2]int{x, y})
			}
		}
	}

	// If no empty spaces found, use default spawn
	if len(emptySpaces) == 0 {
		return 1.5, 1.5
	}

	// Pick random empty space
	chosen := emptySpaces[rand.Intn(len(emptySpaces))]

	// Add some randomness within the cell (0.2 to 0.8 range)
	spawnX := float64(chosen[0]) + 0.2 + rand.Float64()*0.6
	spawnY := float64(chosen[1]) + 0.2 + rand.Float64()*0.6

	return spawnX, spawnY
}

// Update updates the shared game state (projectiles, etc.)
func (gs *GameServer) Update(deltaTime float64) {
	// Update projectiles (thread-safe as it's called from main server loop)
	gs.ProjectileManager.Update(deltaTime, gs.Map)
}

// GetDebugInfo returns debug information about server state
func (gs *GameServer) GetDebugInfo() string {
	gs.PlayersMutex.RLock()
	playerCount := len(gs.Players)
	gs.PlayersMutex.RUnlock()

	// Count active projectiles
	activeProjectiles := 0
	for _, p := range gs.ProjectileManager.Projectiles {
		if p.Active && p.Type == game.Fireball {
			activeProjectiles++
		}
	}

	return fmt.Sprintf("Players: %d/%d | Projectiles: %d",
		playerCount, gs.MaxPlayers, activeProjectiles)
}
