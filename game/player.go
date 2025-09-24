package game

type Player struct {
	Position    Vector
	Direction   Vector
	CameraPlane Vector
	MoveSpeed   float64
	RotSpeed    float64
}

func NewPlayer(x, y float64) *Player {
	return &Player{
		Position:    Vector{x, y},
		Direction:   Vector{-1, 0},   // Initially facing left
		CameraPlane: Vector{0, 0.66}, // FOV of ~60 degrees
		MoveSpeed:   5.0,
		RotSpeed:    3.0,
	}
}

func (p *Player) MoveForward(deltaTime float64, worldMap *Map) {
	newPos := p.Position.Add(p.Direction.Scale(p.MoveSpeed * deltaTime))
	if !worldMap.IsWall(int(newPos.X), int(p.Position.Y)) {
		p.Position.X = newPos.X
	}
	if !worldMap.IsWall(int(p.Position.X), int(newPos.Y)) {
		p.Position.Y = newPos.Y
	}
}

func (p *Player) MoveBackward(deltaTime float64, worldMap *Map) {
	newPos := p.Position.Sub(p.Direction.Scale(p.MoveSpeed * deltaTime))
	if !worldMap.IsWall(int(newPos.X), int(p.Position.Y)) {
		p.Position.X = newPos.X
	}
	if !worldMap.IsWall(int(p.Position.X), int(newPos.Y)) {
		p.Position.Y = newPos.Y
	}
}

func (p *Player) StrafeLeft(deltaTime float64, worldMap *Map) {
	// Perpendicular to direction (rotate 90 degrees counterclockwise)
	strafe := Vector{-p.Direction.Y, p.Direction.X}
	newPos := p.Position.Add(strafe.Scale(p.MoveSpeed * deltaTime))
	if !worldMap.IsWall(int(newPos.X), int(p.Position.Y)) {
		p.Position.X = newPos.X
	}
	if !worldMap.IsWall(int(p.Position.X), int(newPos.Y)) {
		p.Position.Y = newPos.Y
	}
}

func (p *Player) StrafeRight(deltaTime float64, worldMap *Map) {
	// Perpendicular to direction (rotate 90 degrees clockwise)
	strafe := Vector{p.Direction.Y, -p.Direction.X}
	newPos := p.Position.Add(strafe.Scale(p.MoveSpeed * deltaTime))
	if !worldMap.IsWall(int(newPos.X), int(p.Position.Y)) {
		p.Position.X = newPos.X
	}
	if !worldMap.IsWall(int(p.Position.X), int(newPos.Y)) {
		p.Position.Y = newPos.Y
	}
}

func (p *Player) RotateLeft(deltaTime float64) {
	rotSpeed := -p.RotSpeed * deltaTime
	p.Direction = p.Direction.Rotate(rotSpeed)
	p.CameraPlane = p.CameraPlane.Rotate(rotSpeed)
}

func (p *Player) RotateRight(deltaTime float64) {
	rotSpeed := p.RotSpeed * deltaTime
	p.Direction = p.Direction.Rotate(rotSpeed)
	p.CameraPlane = p.CameraPlane.Rotate(rotSpeed)
}
