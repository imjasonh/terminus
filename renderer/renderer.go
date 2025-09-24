package renderer

import (
	"image/color"
	"math"
	"terminus/game"
	"terminus/screen"
)

type Renderer struct {
	screenWidth  int
	screenHeight int
}

func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		screenWidth:  width,
		screenHeight: height,
	}
}

func (r *Renderer) Render(player *game.Player, worldMap *game.Map, screen *screen.Screen, lights []game.LightSource, projectiles []*game.Projectile) {
	screen.Clear()

	// Update renderer to use game area height
	gameHeight := screen.GameHeight

	// Cast rays for each column of the screen
	for x := 0; x < r.screenWidth; x++ {
		// Calculate ray direction
		cameraX := 2*float64(x)/float64(r.screenWidth) - 1 // x-coordinate in camera space
		rayDir := player.Direction.Add(player.CameraPlane.Scale(cameraX))

		// Which box of the map we're in
		mapX := int(player.Position.X)
		mapY := int(player.Position.Y)

		// Length of ray from current position to next x or y side
		var sideDistX, sideDistY float64

		// Length of ray from one x-side to next x-side, or from one y-side to next y-side
		var deltaDistX, deltaDistY float64
		if rayDir.X == 0 {
			deltaDistX = 1e30
		} else {
			deltaDistX = math.Abs(1 / rayDir.X)
		}
		if rayDir.Y == 0 {
			deltaDistY = 1e30
		} else {
			deltaDistY = math.Abs(1 / rayDir.Y)
		}

		var perpWallDist float64

		// What direction to step in x or y-direction (either +1 or -1)
		var stepX, stepY int

		var hit int  // was there a wall hit?
		var side int // was a NS or a EW wall hit?

		// Calculate step and initial sideDist
		if rayDir.X < 0 {
			stepX = -1
			sideDistX = (player.Position.X - float64(mapX)) * deltaDistX
		} else {
			stepX = 1
			sideDistX = (float64(mapX) + 1.0 - player.Position.X) * deltaDistX
		}
		if rayDir.Y < 0 {
			stepY = -1
			sideDistY = (player.Position.Y - float64(mapY)) * deltaDistY
		} else {
			stepY = 1
			sideDistY = (float64(mapY) + 1.0 - player.Position.Y) * deltaDistY
		}

		// Perform DDA
		for hit == 0 {
			// Jump to next map square, either in x-direction, or in y-direction
			if sideDistX < sideDistY {
				sideDistX += deltaDistX
				mapX += stepX
				side = 0
			} else {
				sideDistY += deltaDistY
				mapY += stepY
				side = 1
			}
			// Check if ray has hit a wall
			if worldMap.IsWall(mapX, mapY) {
				hit = 1
			}
		}

		// Calculate distance projected on camera direction
		if side == 0 {
			perpWallDist = (float64(mapX) - player.Position.X + (1-float64(stepX))/2) / rayDir.X
		} else {
			perpWallDist = (float64(mapY) - player.Position.Y + (1-float64(stepY))/2) / rayDir.Y
		}

		// Calculate height of line to draw on screen
		lineHeight := int(float64(gameHeight) / perpWallDist)

		// Calculate lowest and highest pixel to fill in current stripe
		drawStart := -lineHeight/2 + gameHeight/2
		if drawStart < 0 {
			drawStart = 0
		}
		drawEnd := lineHeight/2 + gameHeight/2
		if drawEnd >= gameHeight {
			drawEnd = gameHeight - 1
		}

		// Calculate wall position for lighting
		var wallPos game.Vector
		if side == 0 {
			wallPos = game.Vector{float64(mapX), player.Position.Y + perpWallDist*rayDir.Y}
		} else {
			wallPos = game.Vector{player.Position.X + perpWallDist*rayDir.X, float64(mapY)}
		}

		// Choose wall color based on wall type, side, distance, and lighting
		wallType := worldMap.GetWallType(mapX, mapY)
		wallColor := r.getWallColor(wallType, side, perpWallDist, wallPos, lights)

		// Draw the wall strip
		for y := drawStart; y <= drawEnd; y++ {
			screen.SetCell(x, y, '█', wallColor, wallColor)
		}

		// Draw ceiling with proper distance-based shading
		for y := 0; y < drawStart; y++ {
			// Calculate actual distance to ceiling at this pixel
			// The further from the center line, the further away the ceiling appears
			rowDistance := float64(gameHeight) / (2.0*float64(gameHeight/2-y) - 1.0)
			if rowDistance < 0 {
				rowDistance = perpWallDist // Fallback for edge cases
			}

			ceilingColor := r.getCeilingColor(rowDistance)
			screen.SetCell(x, y, ' ', ceilingColor, ceilingColor)
		}

		// Draw floor with proper distance-based shading
		for y := drawEnd + 1; y < gameHeight; y++ {
			// Calculate actual distance to floor at this pixel
			// The further from the center line, the further away the floor appears
			rowDistance := float64(gameHeight) / (2.0*float64(y-gameHeight/2) - 1.0)
			if rowDistance < 0 {
				rowDistance = perpWallDist // Fallback for edge cases
			}

			floorColor := r.getFloorColor(rowDistance)
			screen.SetCell(x, y, ' ', floorColor, floorColor)
		}
	}

	// Render fireballs as sprites
	r.renderFireballs(player, screen, projectiles)
}

func (r *Renderer) renderFireballs(player *game.Player, screen *screen.Screen, projectiles []*game.Projectile) {
	for _, projectile := range projectiles {
		if !projectile.Active || projectile.Type != game.Fireball {
			continue
		}

		// Transform fireball position relative to player
		relativePos := projectile.Position.Sub(player.Position)

		// Rotate relative to player's view direction
		cos := player.Direction.X
		sin := -player.Direction.Y
		transformedX := cos*relativePos.X - sin*relativePos.Y
		transformedY := sin*relativePos.X + cos*relativePos.Y

		// Skip if behind player
		if transformedY <= 0.1 {
			continue
		}

		// Project to screen coordinates
		screenX := int((float64(r.screenWidth) / 2) * (1.0 + transformedX/transformedY))

		// Check if on screen
		if screenX < 0 || screenX >= r.screenWidth {
			continue
		}

		// Calculate fireball size based on distance - closer = bigger
		gameHeight := screen.GameHeight

		// Use proper perspective projection for size
		fireballSize := int(float64(gameHeight) / transformedY * 0.3) // Scale factor for good visibility

		// Clamp size for reasonable bounds
		if fireballSize < 1 {
			fireballSize = 1 // Very far away = tiny dot
		}
		if fireballSize > gameHeight/2 {
			fireballSize = gameHeight / 2 // Very close = big but not too big
		}

		// Draw fireball
		startY := gameHeight/2 - fireballSize/2
		endY := gameHeight/2 + fireballSize/2

		if startY < 0 {
			startY = 0
		}
		if endY >= gameHeight {
			endY = gameHeight - 1
		}

		// Bright white fireball
		fireballColor := color.RGBA{255, 255, 255, 255} // Pure white

		// Calculate width based on size (bigger fireballs are wider)
		fireballWidth := fireballSize / 3 // Width proportional to height
		if fireballWidth < 1 {
			fireballWidth = 1 // At least 1 pixel wide
		}

		// Draw fireball sprite with proper scaling
		for y := startY; y <= endY; y++ {
			for xOffset := -fireballWidth / 2; xOffset <= fireballWidth/2; xOffset++ {
				drawX := screenX + xOffset
				if drawX >= 0 && drawX < r.screenWidth {
					screen.SetCell(drawX, y, '●', fireballColor, fireballColor)
				}
			}
		}
	}
}

func (r *Renderer) getWallColor(wallType int, side int, distance float64, pos game.Vector, lights []game.LightSource) color.RGBA {
	var baseColor color.RGBA

	switch wallType {
	case 1:
		baseColor = color.RGBA{180, 32, 32, 255} // Dark red walls
	case 2:
		baseColor = color.RGBA{32, 180, 32, 255} // Dark green walls
	case 3:
		baseColor = color.RGBA{32, 32, 180, 255} // Dark blue walls
	case 4:
		baseColor = color.RGBA{180, 180, 32, 255} // Dark yellow walls
	case 5:
		baseColor = color.RGBA{180, 32, 180, 255} // Dark magenta walls
	case 6:
		baseColor = color.RGBA{32, 180, 180, 255} // Cyan walls
	case 7:
		baseColor = color.RGBA{180, 100, 32, 255} // Orange walls
	case 8:
		baseColor = color.RGBA{100, 32, 180, 255} // Purple walls
	default:
		baseColor = color.RGBA{120, 120, 120, 255} // Gray walls
	}

	// Make EW walls darker than NS walls for better depth perception
	sideFactor := 1.0
	if side == 1 {
		sideFactor = 0.7
	}

	// Apply distance-based fog/shading (closer = brighter)
	maxDistance := 8.0 // Objects beyond this distance are very dark
	distanceFactor := 1.0 - (distance / maxDistance)
	if distanceFactor < 0.2 {
		distanceFactor = 0.2 // Minimum visibility
	}

	// Calculate lighting from fireballs
	lightFactor := 0.0
	for _, light := range lights {
		lightFactor += light.GetLightingAt(pos)
	}
	if lightFactor > 1.0 {
		lightFactor = 1.0
	}

	// Combine all factors (distance, side, and lighting)
	finalFactor := sideFactor * (distanceFactor + lightFactor*0.8) // Lighting adds brightness
	if finalFactor > 1.0 {
		finalFactor = 1.0
	}

	return color.RGBA{
		uint8(float64(baseColor.R) * finalFactor),
		uint8(float64(baseColor.G) * finalFactor),
		uint8(float64(baseColor.B) * finalFactor),
		255,
	}
}

func (r *Renderer) getCeilingColor(distance float64) color.RGBA {
	baseColor := color.RGBA{80, 100, 140, 255} // Bluish ceiling

	maxDistance := 10.0
	distanceFactor := 1.0 - (distance / maxDistance)
	if distanceFactor < 0.1 {
		distanceFactor = 0.1
	}

	return color.RGBA{
		uint8(float64(baseColor.R) * distanceFactor),
		uint8(float64(baseColor.G) * distanceFactor),
		uint8(float64(baseColor.B) * distanceFactor),
		255,
	}
}

func (r *Renderer) getFloorColor(distance float64) color.RGBA {
	baseColor := color.RGBA{60, 40, 20, 255} // Brownish floor

	maxDistance := 10.0
	distanceFactor := 1.0 - (distance / maxDistance)
	if distanceFactor < 0.1 {
		distanceFactor = 0.1
	}

	return color.RGBA{
		uint8(float64(baseColor.R) * distanceFactor),
		uint8(float64(baseColor.G) * distanceFactor),
		uint8(float64(baseColor.B) * distanceFactor),
		255,
	}
}
