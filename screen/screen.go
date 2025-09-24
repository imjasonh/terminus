package screen

import (
	"fmt"
	"image/color"
	"strings"
)

type Cell struct {
	Char    rune
	FgColor color.RGBA
	BgColor color.RGBA
}

type Screen struct {
	Width      int
	Height     int
	GameHeight int // Height available for game rendering (excludes HUD)
	Buffer     [][]Cell
	debugMsg   string
}

func NewScreen(width, height int) *Screen {
	buffer := make([][]Cell, height)
	for y := range buffer {
		buffer[y] = make([]Cell, width)
		for x := range buffer[y] {
			buffer[y][x] = Cell{
				Char:    ' ',
				FgColor: color.RGBA{255, 255, 255, 255},
				BgColor: color.RGBA{0, 0, 0, 255},
			}
		}
	}

	return &Screen{
		Width:      width,
		Height:     height,
		GameHeight: height - 2, // Reserve 2 bottom rows for HUD
		Buffer:     buffer,
		debugMsg:   "",
	}
}

func (s *Screen) Clear() {
	// Only clear the game area, not the HUD
	for y := 0; y < s.GameHeight; y++ {
		for x := 0; x < s.Width; x++ {
			s.Buffer[y][x] = Cell{
				Char:    ' ',
				FgColor: color.RGBA{255, 255, 255, 255},
				BgColor: color.RGBA{0, 0, 0, 255},
			}
		}
	}
}

func (s *Screen) SetDebugMessage(msg string) {
	s.debugMsg = msg
}

func (s *Screen) SetCell(x, y int, char rune, fg, bg color.RGBA) {
	// Only allow drawing in the game area, not the HUD area
	if x >= 0 && x < s.Width && y >= 0 && y < s.GameHeight {
		s.Buffer[y][x] = Cell{
			Char:    char,
			FgColor: fg,
			BgColor: bg,
		}
	}
}

func (s *Screen) Render() string {
	var builder strings.Builder

	// Move cursor to top-left and render game area
	builder.WriteString("\x1b[H")

	var lastFg, lastBg color.RGBA
	for y := 0; y < s.GameHeight; y++ {
		// Position cursor at start of this row
		builder.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1))

		for x := 0; x < s.Width; x++ {
			cell := s.Buffer[y][x]

			// Only set colors if they changed (optimization)
			if cell.FgColor != lastFg {
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", cell.FgColor.R, cell.FgColor.G, cell.FgColor.B))
				lastFg = cell.FgColor
			}
			if cell.BgColor != lastBg {
				builder.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm", cell.BgColor.R, cell.BgColor.G, cell.BgColor.B))
				lastBg = cell.BgColor
			}

			builder.WriteRune(cell.Char)
		}
	}

	// Render HUD at bottom
	s.renderHUD(&builder)

	// Reset colors at the end
	builder.WriteString("\x1b[0m")
	return builder.String()
}

func (s *Screen) renderHUD(builder *strings.Builder) {
	// Position cursor at HUD area (second to last row)
	hudRow := s.Height - 1
	fmt.Fprintf(builder, "\x1b[%d;1H", hudRow)

	// Set HUD colors (white text on dark blue background)
	builder.WriteString("\x1b[38;2;255;255;255m\x1b[48;2;0;0;100m")

	// Clear the HUD line and write debug message
	hudLine := fmt.Sprintf("%-*s", s.Width, s.debugMsg)
	if len(hudLine) > s.Width {
		hudLine = hudLine[:s.Width]
	}
	builder.WriteString(hudLine)
}
