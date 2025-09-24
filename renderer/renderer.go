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
	zBuffer      []float64 // Z-buffer for depth testing
}

func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		screenWidth:  width,
		screenHeight: height,
		zBuffer:      make([]float64, width), // Initialize Z-buffer
	}
}

func (r *Renderer) Render(player *game.Player, worldMap *game.Map, screen *screen.Screen, lights []game.LightSource, projectiles []*game.Projectile, otherPlayers []*game.Player) {
	screen.Clear()

	// Clear Z-buffer (initialize with max depth)
	for i := range r.zBuffer {
		r.zBuffer[i] = math.Inf(1) // Infinity represents maximum depth
	}

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
			wallPos = game.Vector{X: float64(mapX), Y: player.Position.Y + perpWallDist*rayDir.Y}
		} else {
			wallPos = game.Vector{X: player.Position.X + perpWallDist*rayDir.X, Y: float64(mapY)}
		}

		// Store wall distance in Z-buffer for sprite depth testing
		r.zBuffer[x] = perpWallDist

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

	// Render all sprites (projectiles and other players)
	r.renderAllSprites(player, screen, projectiles, otherPlayers)
}

func (r *Renderer) renderAllSprites(player *game.Player, screen *screen.Screen, projectiles []*game.Projectile, otherPlayers []*game.Player) {
	// Collect and sort sprites by distance (far to near)
	var sprites []sprite

	// Add projectile sprites
	for _, projectile := range projectiles {
		if !projectile.Active || projectile.Type != game.Fireball {
			continue
		}

		// Transform fireball position relative to player
		relativePos := projectile.Position.Sub(player.Position)

		// Rotate relative to player's view direction using proper 2D rotation
		// We want transformedY to be the distance in front of the player
		transformedY := relativePos.X*player.Direction.X + relativePos.Y*player.Direction.Y
		transformedX := relativePos.X*player.Direction.Y + relativePos.Y*(-player.Direction.X)

		// Skip if behind player
		if transformedY <= 0.1 {
			continue
		}

		sprites = append(sprites, sprite{
			pos:          projectile.Position,
			transformedX: transformedX,
			transformedY: transformedY,
			spriteType:   "fireball",
		})
	}

	// Add other player sprites
	for _, otherPlayer := range otherPlayers {
		// Transform other player position relative to current player
		relativePos := otherPlayer.Position.Sub(player.Position)

		// Rotate relative to player's view direction using proper 2D rotation
		transformedY := relativePos.X*player.Direction.X + relativePos.Y*player.Direction.Y
		transformedX := relativePos.X*player.Direction.Y + relativePos.Y*(-player.Direction.X)

		// Skip if behind player
		if transformedY <= 0.1 {
			continue
		}

		sprites = append(sprites, sprite{
			pos:          otherPlayer.Position,
			transformedX: transformedX,
			transformedY: transformedY,
			spriteType:   "player",
		})
	}

	// Sort sprites from farthest to nearest (painter's algorithm)
	for i := 0; i < len(sprites)-1; i++ {
		for j := i + 1; j < len(sprites); j++ {
			if sprites[i].transformedY < sprites[j].transformedY {
				sprites[i], sprites[j] = sprites[j], sprites[i]
			}
		}
	}

	// Render each sprite
	for _, spr := range sprites {
		r.renderSprite(spr, player, screen)
	}
}

// sprite represents a renderable sprite in 3D space
type sprite struct {
	pos          game.Vector
	transformedX float64
	transformedY float64
	spriteType   string
}

// renderSprite renders a single sprite with proper Z-buffer testing
func (r *Renderer) renderSprite(spr sprite, player *game.Player, screen *screen.Screen) {
	gameHeight := screen.GameHeight

	// Project to screen coordinates using same method as wall renderer
	// Calculate where this sprite appears on screen relative to camera plane
	cameraPlaneLength := math.Sqrt(player.CameraPlane.X*player.CameraPlane.X + player.CameraPlane.Y*player.CameraPlane.Y)
	spriteScreenX := spr.transformedX / spr.transformedY / cameraPlaneLength
	screenX := int(float64(r.screenWidth) / 2 * (1.0 + spriteScreenX))

	// Check if on screen
	if screenX < 0 || screenX >= r.screenWidth {
		return
	}

	// Calculate sprite size based on distance
	var spriteSize int
	var spriteChar rune
	var spriteColor color.RGBA

	switch spr.spriteType {
	case "fireball":
		spriteSize = int(float64(gameHeight) / spr.transformedY * 0.5) // Good size for fireballs
		spriteChar = '●'
		spriteColor = color.RGBA{255, 150, 0, 255} // Bright orange fireball
	case "player":
		// More stable size calculation - less sensitive to small distance changes
		baseSize := float64(gameHeight) / spr.transformedY * 1.2
		spriteSize = int(baseSize + 0.5) // Round properly
		// Clamp to reasonable bounds for stability
		if spriteSize < 4 {
			spriteSize = 4
		}
		spriteChar = '@'
		spriteColor = color.RGBA{0, 255, 0, 255} // Green player
	default:
		return
	}

	// Clamp size
	if spriteSize < 1 {
		spriteSize = 1
	}
	if spriteSize > gameHeight/2 {
		spriteSize = gameHeight / 2
	}

	// Calculate vertical bounds
	startY := gameHeight/2 - spriteSize/2
	endY := gameHeight/2 + spriteSize/2

	if startY < 0 {
		startY = 0
	}
	if endY >= gameHeight {
		endY = gameHeight - 1
	}

	// Calculate horizontal width - make players even wider
	var spriteWidth int
	if spr.spriteType == "player" {
		spriteWidth = (spriteSize * 3) / 4 // Players are much wider - almost as wide as they are tall
	} else {
		spriteWidth = spriteSize / 3 // Fireballs stay normal width
	}
	if spriteWidth < 1 {
		spriteWidth = 1
	}

	// Render sprite with Z-buffer testing
	for xOffset := -spriteWidth / 2; xOffset <= spriteWidth/2; xOffset++ {
		drawX := screenX + xOffset

		// Check bounds and Z-buffer for proper depth testing
		if drawX >= 0 && drawX < r.screenWidth && spr.transformedY < r.zBuffer[drawX]+0.1 {
			// Draw the sprite column
			for y := startY; y <= endY; y++ {
				// Render fireballs with proper appearance
				if spr.spriteType == "fireball" {
					// Simple circular pattern for fireballs
					centerY := startY + (endY-startY)/2
					distFromCenter := math.Abs(float64(y-centerY)) / float64(spriteSize/2+1)
					distFromCenterX := math.Abs(float64(xOffset)) / float64(spriteWidth/2+1)

					intensity := 1.0 - math.Sqrt(distFromCenter*distFromCenter+distFromCenterX*distFromCenterX)
					if intensity > 0.1 { // Low threshold for visibility
						finalColor := color.RGBA{
							uint8(math.Min(255, float64(spriteColor.R)*intensity*1.2)),
							uint8(math.Min(255, float64(spriteColor.G)*intensity*1.2)),
							uint8(math.Min(255, float64(spriteColor.B)*intensity*1.2)),
							255,
						}
						screen.SetCell(drawX, y, spriteChar, finalColor, finalColor)
					}
				} else {
					// Make player sprites more solid and visible
					centerY := startY + (endY-startY)/2
					distFromCenter := math.Abs(float64(y-centerY)) / float64(spriteSize/2+1)
					distFromCenterX := math.Abs(float64(xOffset)) / float64(spriteWidth/2+1)

					intensity := 1.0 - math.Sqrt(distFromCenter*distFromCenter+distFromCenterX*distFromCenterX*0.5) // Less fade on X axis
					if intensity > 0.05 {                                                                           // Very low threshold for maximum visibility
						finalColor := color.RGBA{
							uint8(math.Min(255, float64(spriteColor.R)*intensity*1.5)),
							uint8(math.Min(255, float64(spriteColor.G)*intensity*1.5)),
							uint8(math.Min(255, float64(spriteColor.B)*intensity*1.5)),
							255,
						}
						screen.SetCell(drawX, y, spriteChar, finalColor, finalColor)
					}
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
