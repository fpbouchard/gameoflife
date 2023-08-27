// Game of life
package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
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

var (
	mplusNormalFont font.Face
)

type Game struct {
	// the cells, a 2d array of logicalScreenWidth x logicalScreenHeight of bools
	cells          [][]bool
	pattern        [][]bool
	patternWidth   int
	patternHeight  int
	lastUpdateTime time.Time
	welcomeScreen  bool
	active         bool
	editorVisible  bool
	speed          time.Duration
	terminated     bool
}

func (g *Game) ManageKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.terminated = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.welcomeScreen {
			g.welcomeScreen = false
		} else {
			g.active = !g.active
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.initCells()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		g.initPattern(false)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.editorVisible = !g.editorVisible
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		g.speed -= 25
		if g.speed < 0 {
			g.speed = 0
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		g.speed += 25
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.patternHeight--
		if g.patternHeight < 1 {
			g.patternHeight = 1
		}
		g.initPattern(true)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.patternHeight++
		g.initPattern(true)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.patternWidth++
		g.initPattern(true)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.patternWidth--
		if g.patternWidth < 1 {
			g.patternWidth = 1
		}
		g.initPattern(true)
	}
}

func (g *Game) Update() error {
	timeDelta := time.Since(g.lastUpdateTime)

	g.ManageKeys()
	if g.terminated {
		return ebiten.Termination
	}

	if g.editorVisible && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		patternEditorX := logicalScreenWidth - (g.patternWidth * patternEditorScale)
		patternEditorY := g.patternHeight * patternEditorScale
		if x >= patternEditorX && y <= patternEditorY {
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
			for i := 0; i < g.patternHeight; i++ {
				for j := 0; j < g.patternWidth; j++ {
					g.cells[y+i-g.patternHeight][x+j-g.patternWidth] = g.pattern[i][j]
				}
			}
		}
	}

	if g.active && timeDelta >= g.speed*time.Millisecond {
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
	if g.welcomeScreen {
		// Draw a black background
		vector.DrawFilledRect(screen, 0, 0, logicalScreenWidth, logicalScreenHeight, color.Black, false)
		// Draw the instructions at the center of the screen
		ebitenutil.DebugPrintAt(screen, "Welcome to the Game of Life!\n\n"+
			"Press <space> to start (and pause)\n"+
			"Press <backspace> to clear the screen\n"+
			"Press <delete> to clear the pattern\n"+
			"Press <tab> to show/hide the pattern editor\n"+
			"Press <up>/<down>/<left>/<right> to change the pattern size\n"+
			"Press <+>/<-> to change the speed\n"+
			"Click (and drag) to add the pattern to the screen\n"+
			"Press <escape> to quit\n",
			20, 20,
		)

		return
	}

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

	// Draw pattern editor
	if g.editorVisible {
		patternEditorWidth := g.patternWidth * patternEditorScale
		patternEditorHeight := g.patternHeight * patternEditorScale
		vector.DrawFilledRect(screen, logicalScreenWidth-float32(g.patternWidth)*patternEditorScale, 0, float32(patternEditorWidth), float32(patternEditorHeight), color.Black, false)
		for i := 0; i < g.patternHeight; i++ {
			for j := 0; j < g.patternWidth; j++ {
				x := float32(logicalScreenWidth - ((g.patternWidth - j) * patternEditorScale))
				y := float32(i * patternEditorScale)
				vector.StrokeRect(screen, x, y, patternEditorScale, patternEditorScale, 1, color.RGBA{127, 127, 127, 255}, false)
				if g.pattern[i][j] {
					vector.DrawFilledRect(screen, x, y, patternEditorScale, patternEditorScale, color.White, false)
				}
			}
			// Display the pattern size
			msg := fmt.Sprintf("%dx%d", g.patternWidth, g.patternHeight)
			textBounds := text.BoundString(mplusNormalFont, msg)
			// Center the text under the pattern editor
			text.Draw(screen, msg, mplusNormalFont, (logicalScreenWidth-patternEditorWidth/2)-(textBounds.Dx()/2), patternEditorHeight+10, color.White)
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
	g.cells = make([][]bool, logicalScreenHeight)
	for i := range g.cells {
		g.cells[i] = make([]bool, logicalScreenWidth)
	}
}

func (g *Game) initPattern(keepPrevious bool) {
	newPattern := make([][]bool, g.patternHeight)
	for i := range newPattern {
		newPattern[i] = make([]bool, g.patternWidth)
	}
	if keepPrevious {
		for i := 0; i < g.patternHeight; i++ {
			for j := 0; j < g.patternWidth; j++ {
				if i < len(g.pattern) && j < len(g.pattern[i]) {
					newPattern[i][j] = g.pattern[i][j]
				}
			}
		}
	}
	g.pattern = newPattern
}

func (g *Game) initGlider() {
	g.patternWidth = 3
	g.patternHeight = 3
	g.initPattern(false)
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

func loadFont() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    9,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	loadFont()
	g := &Game{
		lastUpdateTime: time.Now(),
		welcomeScreen:  true,
		active:         false,
		editorVisible:  true,
		patternWidth:   3,
		patternHeight:  3,
		speed:          25,
	}
	g.initCells()
	g.initGlider()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Game of Life")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
