# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Terminus** is a Go-based terminal FPS engine inspired by Wolfenstein 3D that renders 3D scenes using raycasting and ANSI escape codes. The project implements a complete 3D game engine that runs in the terminal with real-time movement, projectile physics, dynamic lighting, and a debug HUD.

## Architecture

### Core Components

**Game Engine (`game/`):**
- `vector.go` - 2D vector math with operations (Add, Sub, Scale, Normalize, Rotate)
- `player.go` - Player state including position, direction, camera plane, and movement methods with collision detection
- `world.go` - Map loading system that reads `.map` files with integer grids (0=empty, 1-8=wall types)
- `projectile.go` - Projectile physics system with fireballs, dynamic lighting, and lifecycle management

**Rendering System (`renderer/`):**
- `renderer.go` - Raycasting engine that projects 3D scenes to 2D using DDA algorithm
  - Wall rendering with distance-based shading and lighting effects
  - Sprite rendering for projectiles with proper 3D positioning
  - Dynamic lighting system that affects wall brightness

**Display System (`screen/`):**
- `screen.go` - Screen buffer and HUD system with ANSI positioning
  - Separates game area from debug HUD (reserves bottom 2 rows)
  - Efficient ANSI rendering that positions cursor instead of scrolling
  - Color management with RGB support

**Main Loop (`main.go`):**
- Terminal size detection and raw input handling
- 30 FPS game loop with delta time calculations
- Map file loading with command-line selection
- Debug information display in HUD

### Rendering Pipeline

1. **Raycasting**: For each screen column, cast a ray from player position through camera plane
2. **DDA Algorithm**: Step ray through map grid until wall intersection
3. **Wall Rendering**: Calculate wall height based on distance, apply lighting and shading
4. **Sprite Rendering**: Project 3D projectile positions to 2D screen coordinates
5. **Screen Output**: Use ANSI positioning to efficiently update display

### Map System

Maps are text files with space-separated integers:
- `0` = empty space
- `1-8` = different wall types with unique colors
- Comments supported with `#`
- Default maps: `maze.map` (tight corridors), `cave.map` (open spaces)

## Development Commands

### Build and Run
```bash
go build                    # Build the executable
./terminus                 # Run with default maze.map
./terminus cave.map        # Run with specific map file
go run main.go maze.map    # Run directly with Go
```

### Map Selection
Maps are loaded at startup via command line argument. The game automatically selects spawn positions based on map type.

## Controls

- `W/A/S/D` - Movement and strafing with collision detection
- `Q/E` - Rotate left/right
- `SPACE` - Shoot fireball projectiles with dynamic lighting
- `ESC` or `Ctrl+C` - Exit

## Key Implementation Details

### Coordinate System
- World coordinates are continuous floats
- Player direction vector defines facing direction
- Camera plane vector (perpendicular to direction) defines FOV (~60 degrees)

### Lighting System
- Fireballs create `LightSource` objects with position, radius, intensity
- Wall colors are modified by distance-based fog and dynamic lighting
- EW walls are rendered darker than NS walls for depth perception

### Screen Management
- Game area uses `screen.GameHeight` (total height - 2 for HUD)
- HUD shows real-time debug info: player position, active projectiles, etc.
- ANSI escape codes used for cursor positioning and true-color support

### Performance
- 30 FPS with delta time for smooth movement
- Efficient raycasting with DDA algorithm
- Optimized ANSI rendering with color change detection