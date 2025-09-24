package game

import (
	"math"
	"math/rand"
)

// NPC represents a non-player character in the game world
type NPC struct {
	Position      Vector
	Direction     Vector
	Speed         float64
	MovementTimer float64 // Time until next direction change
	NPCType       NPCType
}

// NPCType defines different types of NPCs
type NPCType int

const (
	Wanderer NPCType = iota // Basic wandering NPC
)

// NewNPC creates a new NPC at the specified position
func NewNPC(x, y float64, npcType NPCType) *NPC {
	// Random initial direction
	angle := rand.Float64() * 2 * math.Pi
	direction := Vector{math.Cos(angle), math.Sin(angle)}

	return &NPC{
		Position:      Vector{x, y},
		Direction:     direction,
		Speed:         1.5,                      // Slower than players (5.0)
		MovementTimer: 2.0 + rand.Float64()*2.0, // 2-4 seconds until direction change
		NPCType:       npcType,
	}
}

// Update updates the NPC's position and behavior
func (npc *NPC) Update(deltaTime float64, worldMap *Map) {
	// Update movement timer
	npc.MovementTimer -= deltaTime

	// Change direction if timer expired
	if npc.MovementTimer <= 0 {
		npc.changeDirection()
		npc.MovementTimer = 2.0 + rand.Float64()*2.0 // Reset timer for 2-4 seconds
	}

	// Calculate new position
	newPos := npc.Position.Add(npc.Direction.Scale(npc.Speed * deltaTime))

	// Check collision with walls - bounce off if hitting wall
	if worldMap.IsWall(int(newPos.X), int(npc.Position.Y)) {
		// Hit wall horizontally - reverse X direction
		npc.Direction.X = -npc.Direction.X
		npc.MovementTimer = 0.5 // Force direction change soon
	} else {
		npc.Position.X = newPos.X
	}

	if worldMap.IsWall(int(npc.Position.X), int(newPos.Y)) {
		// Hit wall vertically - reverse Y direction
		npc.Direction.Y = -npc.Direction.Y
		npc.MovementTimer = 0.5 // Force direction change soon
	} else {
		npc.Position.Y = newPos.Y
	}

	// Ensure NPC stays within map bounds
	if npc.Position.X < 0.2 || npc.Position.X > float64(worldMap.Width)-0.2 {
		npc.Direction.X = -npc.Direction.X
	}
	if npc.Position.Y < 0.2 || npc.Position.Y > float64(worldMap.Height)-0.2 {
		npc.Direction.Y = -npc.Direction.Y
	}

	// Clamp position to safe bounds
	npc.Position.X = math.Max(0.2, math.Min(float64(worldMap.Width)-0.2, npc.Position.X))
	npc.Position.Y = math.Max(0.2, math.Min(float64(worldMap.Height)-0.2, npc.Position.Y))
}

// changeDirection gives the NPC a new random direction
func (npc *NPC) changeDirection() {
	angle := rand.Float64() * 2 * math.Pi
	npc.Direction = Vector{math.Cos(angle), math.Sin(angle)}
}
