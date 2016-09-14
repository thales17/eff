package eff

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowTitle = "Effulgent"
	frameRate   = 90
	frameTime   = 1000 / frameRate
)

// Point container for 2d points
type Point struct {
	X int
	Y int
}

// Color container for argb colors
type Color struct {
	R int
	G int
	B int
	A int
}

// Canvas interface describing methods required for canvas renderers
type Canvas interface {
	AddDrawable(drawable Drawable)
	Run() int
	DrawPoints(points *[]Point, color Color)
	SetWidth(width int)
	SetHeight(height int)
	GetWidth() int
	GetHeight() int
}

// Drawable interface describing required methods for drawable objects
type Drawable interface {
	Init(canvas Canvas)
	Draw(canvas Canvas)
	Update(canvas Canvas)
}

// SDLCanvas creates window and renderer and calls all drawable methods
type SDLCanvas struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	drawables []Drawable
	width     int
	height    int
}

// SetWidth set the width of the canvas, must be called prior to run
func (sdlCanvas *SDLCanvas) SetWidth(width int) {
	sdlCanvas.width = width
}

// GetWidth get the width of the canvas window
func (sdlCanvas *SDLCanvas) GetWidth() int {
	return sdlCanvas.width
}

// SetHeight set the height of the canvas, must be called prior to run
func (sdlCanvas *SDLCanvas) SetHeight(height int) {
	sdlCanvas.height = height
}

// GetHeight get the height of the canvas window
func (sdlCanvas *SDLCanvas) GetHeight() int {
	return sdlCanvas.height
}

// AddDrawable adds a struct that implements the eff.Drawable interface
func (sdlCanvas *SDLCanvas) AddDrawable(drawable Drawable) {
	sdlCanvas.drawables = append(sdlCanvas.drawables, drawable)
}

// Run creates an infinite loop that renders all drawables, init is only call once and draw and update are called once per frame
func (sdlCanvas *SDLCanvas) Run() int {
	var err error
	sdl.CallQueue <- func() {
		sdlCanvas.window, err = sdl.CreateWindow(
			windowTitle,
			sdl.WINDOWPOS_UNDEFINED,
			sdl.WINDOWPOS_UNDEFINED,
			sdlCanvas.GetWidth(),
			sdlCanvas.GetHeight(),
			sdl.WINDOW_OPENGL,
		)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer func() {
		sdl.CallQueue <- func() {
			sdlCanvas.window.Destroy()
		}
	}()

	sdl.CallQueue <- func() {
		sdlCanvas.renderer, err = sdl.CreateRenderer(
			sdlCanvas.window,
			-1,
			sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC,
		)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create renderer: ", err)
		return 2
	}
	defer func() {
		sdl.CallQueue <- func() {
			sdlCanvas.renderer.Destroy()
		}
	}()

	sdl.CallQueue <- func() {
		sdlCanvas.renderer.Clear()
	}

	// Init Code Goes Here
	for _, drawable := range sdlCanvas.drawables {
		drawable.Init(sdlCanvas)
	}

	running := true
	fullscreen := false
	var lastFrameTime = sdl.GetTicks()
	for running {
		sdl.CallQueue <- func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.KeyUpEvent:
					switch t.Keysym.Sym {
					case sdl.K_q:
						running = false
					case sdl.K_f:
						fullscreen = !fullscreen
						if fullscreen {
							sdlCanvas.window.SetFullscreen(sdl.WINDOW_FULLSCREEN)
						} else {
							sdlCanvas.window.SetFullscreen(0)
						}
					}
				}
			}

			sdlCanvas.renderer.SetDrawColor(0, 0, 0, 0xFF)
			sdlCanvas.renderer.Clear()
		}

		for _, drawable := range sdlCanvas.drawables {
			drawable.Draw(sdlCanvas)
			drawable.Update(sdlCanvas)
		}

		sdl.CallQueue <- func() {
			currentFrameTime := sdl.GetTicks()
			sdlCanvas.renderer.Present()
			if currentFrameTime-lastFrameTime < frameTime {
				sdl.Delay(frameTime - (currentFrameTime - lastFrameTime))
			}
			lastFrameTime = currentFrameTime
		}
	}
	return 0
}

//DrawPoints draw a slice of points to the screen all the same color
func (sdlCanvas *SDLCanvas) DrawPoints(points *[]Point, color Color) {
	sdl.CallQueue <- func() {
		sdlCanvas.renderer.SetDrawColor(
			uint8(color.R),
			uint8(color.G),
			uint8(color.B),
			uint8(color.A),
		)

		sdlPoints := make([]sdl.Point, len(*points))

		for i, point := range *points {
			sdlPoints[i] = sdl.Point{X: int32(point.X), Y: int32(point.Y)}
		}

		sdlCanvas.renderer.DrawPoints(sdlPoints)
	}
}
