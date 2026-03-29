package clock

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

// clockDigits holds 5×7 pixel bitmaps for digits 0–9.
// Each row is a 5-bit value; bit 4 is the leftmost pixel.
var clockDigits = [10][7]uint8{
	{0b01110, 0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b01110}, // 0
	{0b00100, 0b01100, 0b00100, 0b00100, 0b00100, 0b00100, 0b01110}, // 1
	{0b01110, 0b10001, 0b00001, 0b00010, 0b00100, 0b01000, 0b11111}, // 2
	{0b01110, 0b10001, 0b00001, 0b00110, 0b00001, 0b10001, 0b01110}, // 3
	{0b00010, 0b00110, 0b01010, 0b10010, 0b11111, 0b00010, 0b00010}, // 4
	{0b11111, 0b10000, 0b11110, 0b00001, 0b00001, 0b10001, 0b01110}, // 5
	{0b00110, 0b01000, 0b10000, 0b11110, 0b10001, 0b10001, 0b01110}, // 6
	{0b11111, 0b00001, 0b00010, 0b00010, 0b00100, 0b00100, 0b00100}, // 7
	{0b01110, 0b10001, 0b10001, 0b01110, 0b10001, 0b10001, 0b01110}, // 8
	{0b01110, 0b10001, 0b10001, 0b01111, 0b00001, 0b00010, 0b01100}, // 9
}

// clockBrailleBit maps [subRow][subCol] to bit position in the U+2800 braille encoding.
var clockBrailleBit = [4][2]int{
	{0, 3},
	{1, 4},
	{2, 5},
	{6, 7},
}

// colonTop and colonBot are the two braille chars that form a colon separator.
// Dots at pixel rows 2 and 5 of the 7-row digit grid, spanning both columns.
const (
	colonTop = '\u2800' | rune(1<<2|1<<5) // dots at subrow 2 of top char (⠤)
	colonBot = '\u2800' | rune(1<<1|1<<4) // dots at subrow 1 of bottom char (⠒)
)

type Clock struct {
	w, h  int
	theme scene.Theme
	cfg   config.ClockConfig
}

func New(cfg config.ClockConfig) *Clock {
	return &Clock{cfg: cfg}
}

func (c *Clock) Name() string { return "clock" }

func (c *Clock) Init(w, h int, t scene.Theme) {
	c.w, c.h = w, h
	c.theme = t
}

func (c *Clock) Update(_ float64) {}

func (c *Clock) Resize(w, h int) {
	c.w, c.h = w, h
}

func (c *Clock) Draw(screen tcell.Screen) {
	now := time.Now()
	h, m, s := now.Hour(), now.Minute(), now.Second()

	// Each digit: 3 chars wide × 2 chars tall (5×7 pixels in braille).
	// Layout: [H][H] [colon] [M][M] [colon] [S][S]
	// Widths:  3  3  1 1  1   3  3  1 1  1   3  3  = 24 chars total
	const digitW = 3
	const digitH = 2
	const colonW = 1
	const sep = 1 // space on each side of colon

	totalW := 6*digitW + 2*(sep+colonW+sep)
	startX := (c.w - totalW) / 2
	startY := (c.h - digitH) / 2

	mainStyle := c.theme.Bright.Style()
	colonStyle := c.theme.Palette[0].Style()
	dimStyle := c.theme.Dim[0].Style()

	digits := [6]int{h / 10, h % 10, m / 10, m % 10, s / 10, s % 10}

	x := startX
	for i, d := range digits {
		c.drawDigit(screen, d, x, startY, mainStyle)
		x += digitW
		if i == 1 || i == 3 {
			x += sep
			c.setCell(screen, x, startY, colonTop, colonStyle)
			c.setCell(screen, x, startY+1, colonBot, colonStyle)
			x += colonW + sep
		}
	}

	if c.cfg.ShowDate {
		date := now.Format("Monday, January 2")
		dx := (c.w - len(date)) / 2
		dy := startY + digitH + 1
		for i, r := range []rune(date) {
			c.setCell(screen, dx+i, dy, r, dimStyle)
		}
	}
}

// drawDigit renders digit d at character position (cx, cy) in a 3×2 char block.
func (c *Clock) drawDigit(screen tcell.Screen, d, cx, cy int, style tcell.Style) {
	bitmap := clockDigits[d]
	var cells [2][3]rune

	for py := 0; py < 7; py++ {
		for px := 0; px < 5; px++ {
			if (bitmap[py]>>(4-px))&1 == 0 {
				continue
			}
			charCol := px / 2
			charRow := py / 4
			bit := clockBrailleBit[py%4][px%2]
			cells[charRow][charCol] |= 1 << bit
		}
	}

	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			c.setCell(screen, cx+col, cy+row, '\u2800'|cells[row][col], style)
		}
	}
}

func (c *Clock) setCell(screen tcell.Screen, x, y int, r rune, style tcell.Style) {
	if x >= 0 && x < c.w && y >= 0 && y < c.h {
		screen.SetContent(x, y, r, nil, style)
	}
}
