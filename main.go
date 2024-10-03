package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// World represents the game state.
type World struct {
	area   []bool
	width  int
	height int
}

// NewWorld creates a new world.
func NewWorld(width, height int, maxInitLiveCells int) *World {
	w := &World{
		area:   make([]bool, width*height),
		width:  width,
		height: height,
	}
	// w.init(maxInitLiveCells)
	return w
}

const (
	screenWidth  = 320
	screenHeight = 240
)

type Game struct {
	world  *World
	pixels []byte
}

var cnt int = 0

func toInt(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (g *Game) Update() error {
	cnt += 1
	if cnt%60 == 0 {
		log.Printf("Update : %X", opcode)
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	{
		key[0x1] = toInt(ebiten.IsKeyPressed(ebiten.Key1))
		key[0x2] = toInt(ebiten.IsKeyPressed(ebiten.Key2))
		key[0x3] = toInt(ebiten.IsKeyPressed(ebiten.Key3))
		key[0xC] = toInt(ebiten.IsKeyPressed(ebiten.Key4))

		key[0x4] = toInt(ebiten.IsKeyPressed(ebiten.KeyQ))
		key[0x5] = toInt(ebiten.IsKeyPressed(ebiten.KeyW))
		key[0x6] = toInt(ebiten.IsKeyPressed(ebiten.KeyE))
		key[0xD] = toInt(ebiten.IsKeyPressed(ebiten.KeyR))

		key[0x7] = toInt(ebiten.IsKeyPressed(ebiten.KeyA))
		key[0x8] = toInt(ebiten.IsKeyPressed(ebiten.KeyS))
		key[0x9] = toInt(ebiten.IsKeyPressed(ebiten.KeyD))
		key[0xE] = toInt(ebiten.IsKeyPressed(ebiten.KeyF))

		key[0xA] = toInt(ebiten.IsKeyPressed(ebiten.KeyZ))
		key[0x0] = toInt(ebiten.IsKeyPressed(ebiten.KeyX))
		key[0xB] = toInt(ebiten.IsKeyPressed(ebiten.KeyC))
		key[0xF] = toInt(ebiten.IsKeyPressed(ebiten.KeyV))
	}
	EmulateCycle()
	return nil
}

type Coordinate struct {
	X, Y int
}

// generateScaledCoordinates generates a slice of coordinates based on x, y, and scaling factor.
func generateScaledCoordinates(x, y, scaling int) []Coordinate {
	var coordinates []Coordinate

	for i := 0; i < scaling; i++ {
		for j := 0; j < scaling; j++ {
			coordinates = append(coordinates, Coordinate{
				X: scaling*x + i,
				Y: scaling*y + j,
			})
		}
	}

	return coordinates
}

func (g *Game) setPixel(x, y int, color byte) {
	g.pixels[(y*screenWidth+x)*4] = color
	g.pixels[(y*screenWidth+x)*4+1] = color
	g.pixels[(y*screenWidth+x)*4+2] = color
	g.pixels[(y*screenWidth+x)*4+3] = color
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, screenWidth*screenHeight*4)
	}

	if drawFlag {

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				if gfx[(y*64)+x] == 1 {
					for _, coord := range generateScaledCoordinates(x, y, 3) {
						g.setPixel(coord.X, coord.Y, 0xff)
					}
				} else {
					for _, coord := range generateScaledCoordinates(x, y, 3) {
						g.setPixel(coord.X, coord.Y, 0)
					}
				}
			}
		}

		drawFlag = false
	}

	screen.WritePixels(g.pixels)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{
		world: NewWorld(screenWidth, screenHeight, int((screenWidth*screenHeight)/10)),
	}

	Initialize()

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Game of Life (Ebitengine Demo)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
