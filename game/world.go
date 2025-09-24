package game

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Map struct {
	Width  int
	Height int
	Grid   [][]int
}

func NewMap() *Map {
	// Tight maze-like map with narrow corridors and multiple paths
	grid := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 1},
		{1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 1, 2, 1, 0, 1},
		{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1},
		{1, 1, 1, 0, 1, 1, 1, 0, 1, 0, 1, 1, 1, 1, 1, 0, 2, 1, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 2, 0, 0, 1},
		{1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 2, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 0, 1},
		{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 1},
		{1, 0, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 3, 0, 1},
		{1, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 3, 1, 1},
		{1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 3, 0, 1},
		{1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1, 1, 1, 3, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 1, 0, 1, 1},
		{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 4, 0, 0, 4, 1},
		{1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 4, 1, 1, 4, 1},
		{1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	return &Map{
		Width:  20,
		Height: 20,
		Grid:   grid,
	}
}

func (m *Map) IsWall(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return true // Out of bounds is considered a wall
	}
	return m.Grid[y][x] != 0
}

func (m *Map) GetWallType(x, y int) int {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return 1 // Default wall type for out of bounds
	}
	return m.Grid[y][x]
}

func LoadMapFromFile(filename string) (*Map, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var grid [][]int
	var width, height int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Parse each line as space-separated integers
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		row := make([]int, len(parts))
		for i, part := range parts {
			val, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number '%s' in map file: %w", part, err)
			}
			row[i] = val
		}

		grid = append(grid, row)
		if width == 0 {
			width = len(row)
		} else if len(row) != width {
			return nil, fmt.Errorf("inconsistent row width in map file")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading map file: %w", err)
	}

	height = len(grid)
	if height == 0 || width == 0 {
		return nil, fmt.Errorf("empty map file")
	}

	return &Map{
		Width:  width,
		Height: height,
		Grid:   grid,
	}, nil
}
