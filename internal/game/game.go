/*
 * Copyright (c) 2023 Juan Antonio Medina Iglesias
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in
 *  all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 *  THE SOFTWARE.
 */

package game

import (
	"embed"
	"image"
	"image/color"
	"io/fs"
	"math"
	"math/rand"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/juan-medina/classical-concepts-2-trainer/internal/shapes"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	red        = color.RGBA64{0xFFFF, 0x0000, 0x0000, 0xFFFF}
	blue       = color.RGBA64{0x0000, 0x0000, 0xFFFF, 0xFFFF}
	yellow     = color.RGBA64{0xFFFF, 0xFFFF, 0x0000, 0xFFFF}
	purple     = color.RGBA64{0xFFFF, 0x0000, 0xFFFF, 0xFFFF}
	darkPurple = color.RGBA64{0x8888, 0x0000, 0x8888, 0xFFFF}
	darkGreen  = color.RGBA64{0x0000, 0x8888, 0x0000, 0xFFFF}
	green      = color.RGBA64{0x0000, 0xFFFF, 0x0000, 0xFFFF}
	white      = color.RGBA64{0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF}
	gray       = color.RGBA64{0x1111, 0x1111, 0x1111, 0xFFFF}
	lightGray  = color.RGBA64{0x8888, 0x8888, 0x8888, 0xFFFF}
)

const (
	WIDTH         = 1920
	HEIGHT        = 1080
	BUTTON_WIDTH  = 300
	BUTTON_HEIGHT = 100
	NUM_ROWS      = 5
	NUM_COLS      = 7
	TITLE_RADIUS  = 60
	BAR_WIDTH     = 1400
	MAX_TIME      = 15
)

type TileState int

const (
	EmptyTile TileState = iota
	AlphaTile
	BetaTile
	CenterTile
	MouseOverTile
	PlayerTile
	InvalidTile = -1
)

type GameState int

const (
	StandByState GameState = iota
	PlayingState
	EndState
)

type tile struct {
	state    TileState
	x        float32
	y        float32
	rotation float32
}

type BoardPosition struct {
	row, column int
}

type game struct {
	rows                    int
	cols                    int
	board                   [NUM_ROWS][NUM_COLS]tile
	defaultFont             font.Face
	smallFont               font.Face
	aText                   *ebiten.Image
	bText                   *ebiten.Image
	cText                   *ebiten.Image
	dText                   *ebiten.Image
	alphaObjetiveText       *ebiten.Image
	betaObjetiveText        *ebiten.Image
	state                   GameState
	buttonX                 float32
	buttonY                 float32
	buttonOver              bool
	buttonColor             color.Color
	buttonText              *ebiten.Image
	timeLeft                float32
	lastUpdateTime          time.Time
	objectiveX              float32
	objectiveY              float32
	symbolObjective         TileState
	columnObjective         int
	centerSymbolPosition    BoardPosition
	objectiveSymbolPosition BoardPosition
	win                     bool
	winningText             *ebiten.Image
	loosingText             *ebiten.Image
}

func (g game) ShapeHit(shapeX, shapeY float32, pointX, pointY float32) bool {
	var adjustRadius float32 = TITLE_RADIUS * 1.25
	if pointX > shapeX-adjustRadius && pointX < shapeX+adjustRadius && pointY > shapeY-adjustRadius && pointY < shapeY+adjustRadius {
		return true
	}
	return false
}

func (g game) ButtonHit(x, y float32) bool {
	if x > g.buttonX && x < g.buttonX+BUTTON_WIDTH && y > g.buttonY && y < g.buttonY+BUTTON_HEIGHT {
		return true
	}
	return false
}

func (g *game) UpdateButtons() {
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	x, y := ebiten.CursorPosition()
	if g.ButtonHit(float32(x), float32(y)) {
		g.buttonColor = green
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.Reset()
		}
	} else {
		g.buttonColor = darkGreen
	}
}

func (g *game) UpdateTimeBar() {
	// Calculate time elapsed since last update
	elapsedTime := time.Since(g.lastUpdateTime)
	g.lastUpdateTime = time.Now()

	// Convert elapsed time to milliseconds
	elapsedMillis := elapsedTime.Milliseconds()

	// Subtract elapsed time from time left
	g.timeLeft -= float32(elapsedMillis) / 1000 // convert milliseconds to seconds

	if g.timeLeft <= 0 {
		g.timeLeft = 0
		g.End()
	}
}

func (g *game) UpdateBoard() {
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			switch g.board[r][c].state {
			case AlphaTile, BetaTile, CenterTile:
				g.board[r][c].rotation += 1
			}
		}
	}
}

func (g *game) HandleMouseInBoard() {
	x, y := ebiten.CursorPosition()
	cx := float32(x)
	cy := float32(y)

	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			if g.board[r][c].state == EmptyTile || g.board[r][c].state == MouseOverTile {
				if g.ShapeHit(g.board[r][c].x, g.board[r][c].y, cx, cy) {
					ebiten.SetCursorShape(ebiten.CursorShapePointer)
					if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
						g.SetTile(c, r, PlayerTile)
						return
					} else {
						g.SetTile(c, r, MouseOverTile)
						return
					}
				}
			}
		}
	}
}

func (g *game) Update() error {
	switch g.state {
	case StandByState:
		g.UpdateButtons()
	case PlayingState:
		g.UpdateTimeBar()
		g.UpdateBoard()
		g.HandleMouseInBoard()
	case EndState:
		g.UpdateButtons()
	}
	return nil
}

func (g game) DrawButtons(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, g.buttonX, g.buttonY, BUTTON_WIDTH, BUTTON_HEIGHT, g.buttonColor, false)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 0.5)

	op.GeoM.Translate(float64(g.buttonX)+70, float64(g.buttonY)-10)
	screen.DrawImage(g.buttonText, op)

}

func (g game) DrawMarkers(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(150, 0)
	screen.DrawImage(g.aText, op)

	op.GeoM.Translate(360, 0)
	screen.DrawImage(g.bText, op)

	op.GeoM.Translate(360, 0)
	screen.DrawImage(g.cText, op)

	op.GeoM.Translate(360, 0)
	screen.DrawImage(g.dText, op)
}

func (g game) DrawBoard(screen *ebiten.Image) {
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			switch g.board[r][c].state {
			case AlphaTile:
				shapes.DrawPolygon(screen, g.board[r][c].x, g.board[r][c].y, TITLE_RADIUS, 3, g.board[r][c].rotation-90, red)
			case BetaTile:
				shapes.DrawPolygon(screen, g.board[r][c].x, g.board[r][c].y, TITLE_RADIUS, 4, g.board[r][c].rotation-45, yellow)
			case CenterTile:
				shapes.DrawPolygon(screen, g.board[r][c].x, g.board[r][c].y, TITLE_RADIUS, 6, g.board[r][c].rotation, blue)
			case MouseOverTile:
				shapes.DrawPolygon(screen, g.board[r][c].x, g.board[r][c].y, TITLE_RADIUS*1.5, 4, g.board[r][c].rotation-45, lightGray)
			case PlayerTile:
				shapes.DrawPolygon(screen, g.board[r][c].x, g.board[r][c].y, TITLE_RADIUS*1.5, 4, g.board[r][c].rotation-45, white)
			case EmptyTile:
				shapes.DrawPolygon(screen, g.board[r][c].x, g.board[r][c].y, TITLE_RADIUS*1.5, 4, g.board[r][c].rotation-45, gray)
			}
		}
	}
}

func (g game) DrawTimeBar(screen *ebiten.Image) {
	redLength := float32(g.timeLeft) / float32(MAX_TIME) * BAR_WIDTH
	vector.DrawFilledRect(screen, 40, HEIGHT-200, redLength, 100, red, false)
	vector.StrokeRect(screen, 40, HEIGHT-200, BAR_WIDTH, 100, 3, white, false)
}

func (g game) DrawTether(screen *ebiten.Image) {
	center := g.board[g.centerSymbolPosition.row][g.centerSymbolPosition.column]
	objective := g.board[g.objectiveSymbolPosition.row][g.objectiveSymbolPosition.column]

	var fromX, fromY, width, height float32

	// if is horizontal draw a horizontal line using a rect
	if objective.y == center.y {
		if objective.x < center.x {
			fromX = objective.x
			fromY = objective.y
		} else {
			fromX = center.x
			fromY = center.y
		}
		width = float32(math.Abs(float64(objective.x - center.x)))
		height = 5
	} else {
		if objective.y < center.y {
			fromX = objective.x
			fromY = objective.y
		} else {
			fromX = center.x
			fromY = center.y
		}
		width = 5
		height = float32(math.Abs(float64(objective.y - center.y)))
	}

	vector.DrawFilledRect(screen, fromX, fromY, width, height, darkPurple, false)
}

func (g game) DrawObjective(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(float64(g.objectiveX), float64(g.objectiveY))

	var objectiveColor color.RGBA64
	switch g.columnObjective {
	case 0:
		objectiveColor = red
	case 1:
		objectiveColor = green
	case 2:
		objectiveColor = blue
	case 3:
		objectiveColor = purple
	}

	rc, gc, bc, ac := objectiveColor.RGBA()

	op.ColorScale.Scale(float32(rc)/float32(255), float32(gc)/float32(255), float32(bc)/float32(255), float32(ac)/float32(255))

	if g.symbolObjective == AlphaTile {
		screen.DrawImage(g.alphaObjetiveText, op)
	} else {
		screen.DrawImage(g.betaObjetiveText, op)
	}

}

func (g game) DrawWinningStatus(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(500, 900)
	if g.win {
		screen.DrawImage(g.winningText, op)
	} else {
		screen.DrawImage(g.loosingText, op)
	}
}

func (g game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StandByState:
		g.DrawButtons(screen)
		g.DrawMarkers(screen)
	case PlayingState:
		g.DrawBoard(screen)
		g.DrawTimeBar(screen)
		g.DrawMarkers(screen)
		g.DrawObjective(screen)
	case EndState:
		g.DrawButtons(screen)
		g.DrawBoard(screen)
		g.DrawMarkers(screen)
		g.DrawObjective(screen)
		g.DrawTether(screen)
		g.DrawWinningStatus(screen)
	}
}

func (g game) CreateTextImage(text string, color color.Color, face font.Face) *ebiten.Image {
	textImage := image.NewRGBA(g.getTextDimensions(text))

	// Draw the text on the image
	drawer := &font.Drawer{
		Dst:  textImage,
		Src:  image.NewUniform(color),
		Face: face,
		Dot:  fixed.P(0, int(face.Metrics().Height.Ceil())),
	}
	drawer.DrawString(text)

	// Convert *image.RGBA to *ebiten.Image
	return ebiten.NewImageFromImage(textImage)
}

func (g game) getTextDimensions(text string) image.Rectangle {
	width := 0
	maxHeight := 0
	minHeight := 0

	for _, ch := range text {
		b, a, ok := g.defaultFont.GlyphBounds(ch)
		if !ok {
			continue
		}
		if int(b.Max.Y) > maxHeight {
			maxHeight = int(b.Max.Y)
		}
		if int(b.Min.Y) < minHeight {
			minHeight = int(b.Min.Y)
		}
		width += a.Ceil()
	}

	height := maxHeight - minHeight
	return image.Rect(0, 0, width, height)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WIDTH, HEIGHT
}

func (g *game) Standby() {
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			g.board[r][c].state = InvalidTile
		}
	}
	g.state = StandByState
}

func (g *game) End() {

	var states [NUM_ROWS][NUM_COLS]TileState

	g.RemoveTileWithState(MouseOverTile)

	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			currentState := g.board[r][c].state
			if currentState == AlphaTile || currentState == BetaTile || currentState == CenterTile {
				// flip horizontally r
				nr := g.rows - 1 - r
				// flip vertically c
				nc := g.cols - 1 - c
				states[nr][nc] = currentState
			}
		}
	}

	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			currentState := states[r][c]
			if currentState == AlphaTile || currentState == BetaTile || currentState == CenterTile {
				g.board[r][c].state = currentState
			}
		}
	}

	objectiveRow := 0
	objectiveColumn := g.columnObjective * 2

	for r := 0; r < g.rows; r++ {
		if g.board[r][objectiveColumn].state == CenterTile {
			objectiveRow = r
			break
		}
	}

	var objectivePosition BoardPosition
	found := false

	posible := g.TilesAroundATileWithAnState(objectiveRow, objectiveColumn, g.symbolObjective)
	possibles := len(posible)
	if possibles > 0 {
		if possibles == 1 {
			found = true
			objectivePosition = posible[0]
		} else {
			for _, p := range posible {
				if len(g.TilesAroundATileWithAnState(p.row, p.column, CenterTile)) == 1 {
					found = true
					objectivePosition = p
					break
				}
			}
		}

	}

	if found {
		g.objectiveSymbolPosition = objectivePosition
		g.centerSymbolPosition = BoardPosition{row: objectiveRow, column: objectiveColumn}
		playerFound, playerPosition := g.FindPlayerPosition()
		if playerFound {
			// check is between the center and objective

			// if he need to be horizontally
			if objectivePosition.row == g.centerSymbolPosition.row && g.centerSymbolPosition.row == playerPosition.row {
				if objectivePosition.column < g.centerSymbolPosition.column {
					if playerPosition.column > objectivePosition.column && playerPosition.column < g.centerSymbolPosition.column {
						g.objectiveSymbolPosition = playerPosition
						g.win = true
					}
				} else {
					if playerPosition.column < objectivePosition.column && playerPosition.column > g.centerSymbolPosition.column {
						g.objectiveSymbolPosition = playerPosition
						g.win = true
					}
				}
			} else if objectivePosition.column == g.centerSymbolPosition.column && g.centerSymbolPosition.column == playerPosition.column {
				if objectivePosition.row < g.centerSymbolPosition.row && g.centerSymbolPosition.row < playerPosition.row {
					if playerPosition.row > objectivePosition.row && playerPosition.row < g.centerSymbolPosition.row {
						g.objectiveSymbolPosition = playerPosition
						g.win = true
					}
				} else {
					if playerPosition.row < objectivePosition.row && playerPosition.row > g.centerSymbolPosition.row {
						g.objectiveSymbolPosition = playerPosition
						g.win = true
					}
				}
			}
		}
	}

	g.state = EndState
}

func (g *game) FindPlayerPosition() (bool, BoardPosition) {
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			if g.board[r][c].state == PlayerTile {
				return true, BoardPosition{row: r, column: c}
			}
		}
	}
	return false, BoardPosition{}
}

func (g game) TilesAroundATileWithAnState(row, column int, state TileState) []BoardPosition {
	result := []BoardPosition{}

	// look 2 up
	if row > 1 {
		if g.board[row-2][column].state == state {
			result = append(result, BoardPosition{row: row - 2, column: column})
		}
	}
	// look 2 down
	if row < g.rows-2 {
		if g.board[row+2][column].state == state {
			result = append(result, BoardPosition{row: row + 2, column: column})
		}
	}
	// look 2 left
	if column > 1 {
		if g.board[row][column-2].state == state {
			result = append(result, BoardPosition{row: row, column: column - 2})
		}
	}
	// look 2 right
	if column < g.cols-2 {
		if g.board[row][column+2].state == state {
			result = append(result, BoardPosition{row: row, column: column + 2})
		}
	}

	return result
}

func (g *game) Reset() {
	const (
		startX = TITLE_RADIUS * 3
		startY = TITLE_RADIUS * 3
	)

	var x, y float32 = startX, startY
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			g.board[r][c].state = EmptyTile

			g.board[r][c].rotation = 0
			g.board[r][c].x = x
			g.board[r][c].y = y
			x += TITLE_RADIUS * 3
		}
		x = startX
		y += TITLE_RADIUS * 2.5
	}

	g.board[1][1].state = InvalidTile
	g.board[3][1].state = InvalidTile
	g.board[1][3].state = InvalidTile
	g.board[3][3].state = InvalidTile
	g.board[1][5].state = InvalidTile
	g.board[3][5].state = InvalidTile

	g.SetTile(0, 0, BetaTile)
	g.SetTile(0, 2, CenterTile)
	g.SetTile(0, 4, AlphaTile)

	g.SetTile(2, 0, CenterTile)
	g.SetTile(2, 2, AlphaTile)
	g.SetTile(2, 4, BetaTile)

	g.SetTile(4, 0, BetaTile)
	g.SetTile(4, 2, BetaTile)
	g.SetTile(4, 4, CenterTile)

	g.SetTile(6, 0, AlphaTile)
	g.SetTile(6, 2, CenterTile)
	g.SetTile(6, 4, AlphaTile)

	g.state = PlayingState
	g.timeLeft = MAX_TIME
	g.lastUpdateTime = time.Now()

	// random alpha or beta
	g.symbolObjective = TileState(rand.Intn(2) + 1)

	// random 0, 1, 2, 3
	g.columnObjective = rand.Intn(4)
	g.win = false
}

func (g *game) RemoveTileWithState(state TileState) {
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			if g.board[r][c].state == state {
				g.board[r][c].state = EmptyTile
			}
		}
	}
}

func (g *game) SetTile(c int, r int, state TileState) {
	switch state {
	case PlayerTile, MouseOverTile:
		g.RemoveTileWithState(state)

	}

	g.board[r][c].state = state
	g.board[r][c].rotation = 0
}

func New(er embed.FS) ebiten.Game {
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowTitle("Classical Concepts 2 Trainer")
	ebiten.SetTPS(60)

	// Load font
	fontBytes, err := fs.ReadFile(er, "embed/fonts/default.ttf")
	if err != nil {
		panic(err)
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}

	defaultFont := truetype.NewFace(font, &truetype.Options{
		Size: 70,
		DPI:  90,
	})

	smallFont := truetype.NewFace(font, &truetype.Options{
		Size: 50,
		DPI:  90,
	})

	g := game{
		board:       [NUM_ROWS][NUM_COLS]tile{},
		rows:        NUM_ROWS,
		cols:        NUM_COLS,
		defaultFont: defaultFont,
		smallFont:   smallFont,
	}

	g.Standby()

	g.buttonX = WIDTH - (BUTTON_WIDTH * 1.5)
	g.buttonY = (HEIGHT / 2) - (BUTTON_HEIGHT / 2)
	g.buttonOver = false
	g.buttonColor = darkGreen

	g.objectiveX = WIDTH - 400
	g.objectiveY = 100

	g.aText = g.CreateTextImage("A", red, g.defaultFont)
	g.bText = g.CreateTextImage("B", yellow, g.defaultFont)
	g.cText = g.CreateTextImage("C", blue, g.defaultFont)
	g.dText = g.CreateTextImage("D", purple, g.defaultFont)

	g.buttonText = g.CreateTextImage("Try!", white, g.defaultFont)

	g.alphaObjetiveText = g.CreateTextImage("Alpha", white, g.defaultFont)
	g.betaObjetiveText = g.CreateTextImage("Beta", white, g.defaultFont)

	g.winningText = g.CreateTextImage("Great Success!", green, g.defaultFont)
	g.loosingText = g.CreateTextImage("Oh, my bad!", red, g.defaultFont)

	return &g
}
