# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Terminus** is a multiplayer SSH-based terminal FPS engine inspired by Wolfenstein 3D that renders 3D scenes using raycasting and ANSI escape codes. The project implements a complete multiplayer 3D game engine accessible via SSH, featuring real-time player interaction, wandering NPCs, projectile physics, dynamic lighting, and a debug HUD.

## Architecture

### Core Components

**Game Engine (`game/`):**
- `vector.go` - 2D vector math with operations (Add, Sub, Scale, Normalize, Rotate)
- `player.go` - Player state including position, direction, camera plane, and movement methods with collision detection
- `world.go` - Map loading system that reads `.map` files with integer grids (0=empty, 1-8=wall types)
- `projectile.go` - Projectile physics system with fireballs, dynamic lighting, and lifecycle management
- `npc.go` - NPC system with random walk AI, collision detection, and wandering behavior

**Rendering System (`renderer/`):**
- `renderer.go` - Raycasting engine that projects 3D scenes to 2D using DDA algorithm
  - Wall rendering with distance-based shading and lighting effects
  - Advanced sprite rendering for projectiles, players, and NPCs with Z-buffer depth testing
  - Dynamic lighting system that affects wall brightness
  - Proper sprite sorting and perspective projection for multiplayer visibility

**Display System (`screen/`):**
- `screen.go` - Screen buffer and HUD system with ANSI positioning
  - Separates game area from debug HUD (reserves bottom 2 rows)
  - Efficient ANSI rendering that positions cursor instead of scrolling
  - Color management with RGB support

**SSH Server & Main Loop (`main.go`):**
- SSH server on port 2222 with persistent host key generation
- Per-player game session management with goroutines
- Terminal size detection from SSH PTY and input handling
- 30 FPS shared game loop with delta time calculations
- Map file loading with command-line selection

**Multiplayer Server (`server/`):**
- `server.go` - GameServer with thread-safe player and NPC management
  - Shared world state with up to 10 concurrent players
  - Random spawn point generation for players and NPCs
  - NPC spawning and lifecycle management (3-5 NPCs per map)

### Rendering Pipeline

1. **Raycasting**: For each screen column, cast a ray from player position through camera plane
2. **DDA Algorithm**: Step ray through map grid until wall intersection
3. **Wall Rendering**: Calculate wall height based on distance, apply lighting and shading
4. **Z-Buffer Population**: Store wall distances for proper sprite depth testing
5. **Sprite Collection**: Gather all sprites (projectiles, other players, NPCs) visible to current player
6. **Sprite Sorting**: Sort sprites by distance (painter's algorithm)
7. **Sprite Rendering**: Project 3D sprite positions to 2D with Z-buffer testing for proper occlusion
8. **Screen Output**: Use ANSI positioning to efficiently update display

### Map System

Maps are text files with space-separated integers:
- `0` = empty space
- `1-8` = different wall types with unique colors
- Comments supported with `#`
- Default maps: `maze.map` (tight corridors), `cave.map` (open spaces)

## Development Commands

### Build and Run SSH Server
```bash
go build                    # Build the SSH server
./terminus                 # Start SSH server with default maze.map on port 2222
./terminus cave.map        # Start SSH server with cave.map
go run . cave.map          # Run SSH server directly with Go
```

### Connect to Server
```bash
ssh -p 2222 localhost      # Connect to local server
```

### Map Selection
Maps are loaded at server startup via command line argument. Players spawn at random empty locations.

## Controls (per SSH client)

- `W/A/S/D` - Movement and strafing with collision detection
- `Q/E` - Rotate left/right
- `SPACE` - Shoot fireball projectiles with dynamic lighting (visible to all players)
- `ESC` or `Ctrl+C` - Exit

## Key Implementation Details

### Coordinate System
- World coordinates are continuous floats
- Player direction vector defines facing direction
- Camera plane vector (perpendicular to direction) defines FOV (~60 degrees)

### Multiplayer Architecture
- **SSH Server**: Handles up to 10 concurrent connections on port 2222
- **Per-Player Sessions**: Each SSH connection gets isolated game loop goroutine
- **Shared State**: Map, projectiles, and NPCs shared across all players
- **Thread Safety**: Mutex protection for concurrent access to shared data

### Sprite System
- **Players**: Large green `@` symbols (1.2x scale, 75% width-to-height ratio)
- **NPCs**: Medium blue `◐` symbols (1.0x scale, 50% width) with random walk AI
- **Projectiles**: Orange `●` symbols (0.5x scale) with circular fade patterns
- **Z-Buffer Testing**: Proper depth testing so sprites hide behind walls
- **Coordinate Transformation**: Proper 3D-to-2D projection using camera plane

### Lighting System
- Fireballs create `LightSource` objects with position, radius, intensity
- Wall colors are modified by distance-based fog and dynamic lighting
- EW walls are rendered darker than NS walls for depth perception

### Screen Management
- Game area uses `screen.GameHeight` (total height - 2 for HUD)
- HUD shows real-time debug info: player position, player count, active projectiles
- ANSI escape codes used for cursor positioning and true-color support
- Per-player rendering with terminal resize support

### Performance
- 30 FPS server-side game loop with delta time for smooth movement
- 30 FPS per-player rendering loops
- Efficient raycasting with DDA algorithm
- Optimized ANSI rendering with color change detection
- Thread-safe concurrent player and NPC updates
