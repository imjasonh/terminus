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
	Width  int
	Height int
	Buffer [][]Cell
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
		Width:  width,
		Height: height,
		Buffer: buffer,
	}
}

func (s *Screen) Clear() {
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			s.Buffer[y][x] = Cell{
				Char:    ' ',
				FgColor: color.RGBA{255, 255, 255, 255},
				BgColor: color.RGBA{0, 0, 0, 255},
			}
		}
	}
}

func (s *Screen) SetCell(x, y int, char rune, fg, bg color.RGBA) {
	if x >= 0 && x < s.Width && y >= 0 && y < s.Height {
		s.Buffer[y][x] = Cell{
			Char:    char,
			FgColor: fg,
			BgColor: bg,
		}
	}
}

func (s *Screen) Render() string {
	var builder strings.Builder
	var lastFg, lastBg color.RGBA

	// Move cursor to top-left
	builder.WriteString("\x1b[H")

	for y := 0; y < s.Height; y++ {
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

	// Reset colors at the end
	builder.WriteString("\x1b[0m")
	return builder.String()
}
