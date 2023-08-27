// Game of life
package main

import (
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth         = 1280
	screenHeight        = 960
	logicalScreenFactor = 2
	logicalScreenWidth  = screenWidth / logicalScreenFactor
	logicalScreenHeight = screenHeight / logicalScreenFactor
	cellSize            = 2
	patternEditorScale  = 20
)

type Game struct {
	// the cells, a 2d array of logicalScreenWidth x logicalScreenHeight of bools
	cells          [][]bool
	pattern        [][]bool
	patternWidth   int
	patternHeight  int
	lastUpdateTime time.Time
	active         bool
	editorVisible  bool
}

func (g *Game) Update() error {
	timeDelta := time.Since(g.lastUpdateTime)

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.initCells()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.initPattern()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.active = !g.active
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.editorVisible = !g.editorVisible
	}

	if g.editorVisible && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		patternEditorX := logicalScreenWidth - (g.patternWidth * patternEditorScale)
		patternEditorY := g.patternHeight * patternEditorScale
		// If the cursor is within the pattern editor, toggle the pattern pixel
		if x >= patternEditorX && y <= patternEditorY {
			// Translate the cursor position to the pattern editor (0,0) is the top left corner, (patternSize - 1, patternSize - 1) is the bottom right corner
			x = (x - patternEditorX) / patternEditorScale
			y = y / patternEditorScale
			g.pattern[y][x] = !g.pattern[y][x]
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		patternEditorX := logicalScreenWidth - (g.patternWidth * patternEditorScale)
		patternEditorY := g.patternHeight * patternEditorScale
		if !g.editorVisible || x < patternEditorX || y > patternEditorY {
			x = x / logicalScreenFactor
			y = y / logicalScreenFactor
			// Print the pattern at the cursor position (aligned top-left)
			for i := 0; i < g.patternHeight; i++ {
				for j := 0; j < g.patternWidth; j++ {
					if g.pattern[i][j] {
						g.cells[y+i-g.patternHeight][x+j-g.patternWidth] = true
					}
				}
			}
		}
	}

	// Apply rules of life once every 10 frames
	if g.active && timeDelta >= 25*time.Millisecond {
		g.lastUpdateTime = time.Now()

		// Rules of life:
		// 1. Any live cell with fewer than two live neighbours dies, as if by underpopulation.
		// 2. Any live cell with two or three live neighbours lives on to the next generation.
		// 3. Any live cell with more than three live neighbours dies, as if by overpopulation.
		// 4. Any dead cell with exactly three live neighbours becomes a live cell, as if by reproduction.

		// Make a copy of the cells
		cellsCopy := make([][]bool, logicalScreenHeight)
		for i := range cellsCopy {
			cellsCopy[i] = make([]bool, logicalScreenWidth)
		}

		for i := 0; i < logicalScreenHeight; i++ {
			for j := 0; j < logicalScreenWidth; j++ {
				// Count the number of live neighbors
				neighbors := 0
				for y := i - 1; y <= i+1; y++ {
					for x := j - 1; x <= j+1; x++ {
						if x >= 0 && x < logicalScreenWidth && y >= 0 && y < logicalScreenHeight && !(x == j && y == i) {
							if g.cells[y][x] {
								neighbors++
							}
						}
					}
				}
				// Apply the rules of life
				if g.cells[i][j] {
					if neighbors < 2 || neighbors > 3 {
						cellsCopy[i][j] = false
					} else {
						cellsCopy[i][j] = true
					}
				} else {
					if neighbors == 3 {
						cellsCopy[i][j] = true
					} else {
						cellsCopy[i][j] = false
					}
				}
			}
		}
		// Copy the cells back
		g.cells = cellsCopy
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw a blue square at the cursor position
	x, y := ebiten.CursorPosition()
	vector.DrawFilledRect(screen, float32(x-cellSize/2), float32(y-cellSize/2), cellSize, cellSize, color.RGBA{0, 0, 255, 255}, false)
	// Draw a pixel at each cell that is alive
	for i := 0; i < logicalScreenHeight; i++ {
		for j := 0; j < logicalScreenWidth; j++ {
			if g.cells[i][j] {
				x := float32(j * logicalScreenFactor)
				y := float32(i * logicalScreenFactor)
				vector.DrawFilledRect(screen, x, y, cellSize, cellSize, color.White, false)
			}
		}
	}

	if g.editorVisible {
		vector.DrawFilledRect(screen, logicalScreenWidth-float32(g.patternWidth)*patternEditorScale, 0, logicalScreenWidth, float32(g.patternHeight)*patternEditorScale, color.Black, false)
		for i := 0; i < g.patternHeight; i++ {
			for j := 0; j < g.patternWidth; j++ {
				x := float32(logicalScreenWidth - ((g.patternWidth - j) * patternEditorScale))
				y := float32(i * patternEditorScale)
				vector.StrokeRect(screen, x, y, patternEditorScale, patternEditorScale, 1, color.White, false)
				if g.pattern[i][j] {
					vector.DrawFilledRect(screen, x, y, patternEditorScale, patternEditorScale, color.White, false)
				}
			}
		}

	}

	// In the lower right corner, draw a pause symbol if the game is paused (two rectangles)
	if !g.active {
		vector.DrawFilledRect(screen, logicalScreenWidth-60, logicalScreenHeight-80, 10, 40, color.White, false)
		vector.DrawFilledRect(screen, logicalScreenWidth-40, logicalScreenHeight-80, 10, 40, color.White, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return logicalScreenWidth, logicalScreenHeight
}

func (g *Game) initCells() {
	// Reinitialize cells array
	g.cells = make([][]bool, logicalScreenHeight)
	for i := range g.cells {
		g.cells[i] = make([]bool, logicalScreenWidth)
	}
}

func (g *Game) initPattern() {
	// Reinitialize cells array
	g.pattern = make([][]bool, g.patternHeight)
	for i := range g.pattern {
		g.pattern[i] = make([]bool, g.patternWidth)
	}
}

func (g *Game) initGlider() {
	g.patternWidth = 3
	g.patternHeight = 3
	g.initPattern()
	// Build a basic glider pattern
	// 0 1 0
	// 0 0 1
	// 1 1 1
	g.pattern[0][1] = true
	g.pattern[1][2] = true
	g.pattern[2][0] = true
	g.pattern[2][1] = true
	g.pattern[2][2] = true
}

func main() {
	// Initialize the Game struct
	g := &Game{
		lastUpdateTime: time.Now(),
		active:         true,
		editorVisible:  true,
		patternWidth:   3,
		patternHeight:  3,
	}
	g.initCells()
	g.initGlider()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
