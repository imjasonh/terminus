package game

type Projectile struct {
	Position  Vector
	Direction Vector
	Speed     float64
	Life      float64 // Time to live in seconds
	MaxLife   float64
	Active    bool
	Type      ProjectileType
}

type ProjectileType int

const (
	Fireball ProjectileType = iota
)

func NewFireball(startPos, direction Vector) *Projectile {
	return &Projectile{
		Position:  startPos,
		Direction: direction.Normalize(),
		Speed:     8.0, // Units per second
		Life:      3.0, // 3 seconds to live
		MaxLife:   3.0,
		Active:    true,
		Type:      Fireball,
	}
}

func (p *Projectile) Update(deltaTime float64, worldMap *Map) {
	if !p.Active {
		return
	}

	// Update lifetime
	p.Life -= deltaTime
	if p.Life <= 0 {
		p.Active = false
		return
	}

	// Calculate new position
	movement := p.Direction.Scale(p.Speed * deltaTime)
	newPos := p.Position.Add(movement)

	// Check for wall collision
	if worldMap.IsWall(int(newPos.X), int(newPos.Y)) {
		p.Active = false
		return
	}

	p.Position = newPos
}

func (p *Projectile) GetLightRadius() float64 {
	if !p.Active || p.Type != Fireball {
		return 0
	}

	// Light radius changes over lifetime (brighter when fresh)
	lifeRatio := p.Life / p.MaxLife
	return 2.0 + 1.5*lifeRatio // Radius from 2.0 to 3.5
}

func (p *Projectile) GetLightIntensity() float64 {
	if !p.Active || p.Type != Fireball {
		return 0
	}

	// Intensity fades over lifetime
	lifeRatio := p.Life / p.MaxLife
	return 0.8 * lifeRatio // Intensity from 0 to 0.8
}

type ProjectileManager struct {
	Projectiles []*Projectile
}

func NewProjectileManager() *ProjectileManager {
	return &ProjectileManager{
		Projectiles: make([]*Projectile, 0),
	}
}

func (pm *ProjectileManager) AddProjectile(p *Projectile) {
	pm.Projectiles = append(pm.Projectiles, p)
}

func (pm *ProjectileManager) Update(deltaTime float64, worldMap *Map) {
	// Update all projectiles
	for _, p := range pm.Projectiles {
		p.Update(deltaTime, worldMap)
	}

	// Remove inactive projectiles
	activeProjectiles := make([]*Projectile, 0)
	for _, p := range pm.Projectiles {
		if p.Active {
			activeProjectiles = append(activeProjectiles, p)
		}
	}
	pm.Projectiles = activeProjectiles
}

func (pm *ProjectileManager) GetActiveLights() []LightSource {
	lights := make([]LightSource, 0)
	for _, p := range pm.Projectiles {
		if p.Active && p.GetLightRadius() > 0 {
			lights = append(lights, LightSource{
				Position:  p.Position,
				Radius:    p.GetLightRadius(),
				Intensity: p.GetLightIntensity(),
				Color:     [3]float64{1.0, 0.6, 0.2}, // Orange-red fireball light
			})
		}
	}
	return lights
}

type LightSource struct {
	Position  Vector
	Radius    float64
	Intensity float64
	Color     [3]float64 // RGB values 0-1
}

func (ls LightSource) GetLightingAt(pos Vector) float64 {
	distance := pos.Sub(ls.Position).Length()
	if distance > ls.Radius {
		return 0
	}

	// Smooth falloff
	falloff := 1.0 - (distance / ls.Radius)
	return ls.Intensity * falloff * falloff // Quadratic falloff
}
