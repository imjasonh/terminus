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

func (r *Renderer) Render(player *game.Player, worldMap *game.Map, screen *screen.Screen) {
	screen.Clear()

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
		lineHeight := int(float64(r.screenHeight) / perpWallDist)

		// Calculate lowest and highest pixel to fill in current stripe
		drawStart := -lineHeight/2 + r.screenHeight/2
		if drawStart < 0 {
			drawStart = 0
		}
		drawEnd := lineHeight/2 + r.screenHeight/2
		if drawEnd >= r.screenHeight {
			drawEnd = r.screenHeight - 1
		}

		// Choose wall color based on wall type, side, and distance
		wallType := worldMap.GetWallType(mapX, mapY)
		wallColor := r.getWallColor(wallType, side, perpWallDist)

		// Draw the wall strip
		for y := drawStart; y <= drawEnd; y++ {
			screen.SetCell(x, y, '█', wallColor, wallColor)
		}

		// Draw ceiling with proper distance-based shading
		for y := 0; y < drawStart; y++ {
			// Calculate actual distance to ceiling at this pixel
			// The further from the center line, the further away the ceiling appears
			rowDistance := float64(r.screenHeight) / (2.0*float64(r.screenHeight/2-y) - 1.0)
			if rowDistance < 0 {
				rowDistance = perpWallDist // Fallback for edge cases
			}

			ceilingColor := r.getCeilingColor(rowDistance)
			screen.SetCell(x, y, ' ', ceilingColor, ceilingColor)
		}

		// Draw floor with proper distance-based shading
		for y := drawEnd + 1; y < r.screenHeight; y++ {
			// Calculate actual distance to floor at this pixel
			// The further from the center line, the further away the floor appears
			rowDistance := float64(r.screenHeight) / (2.0*float64(y-r.screenHeight/2) - 1.0)
			if rowDistance < 0 {
				rowDistance = perpWallDist // Fallback for edge cases
			}

			floorColor := r.getFloorColor(rowDistance)
			screen.SetCell(x, y, ' ', floorColor, floorColor)
		}
	}
}

func (r *Renderer) getWallColor(wallType int, side int, distance float64) color.RGBA {
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

	// Combine both factors
	finalFactor := sideFactor * distanceFactor

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
