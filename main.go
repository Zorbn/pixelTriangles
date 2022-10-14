package main

import (
	"log"
	"runtime"

	"golang.org/x/exp/constraints"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	screenWidth    = 320
	screenHeight   = 180
	screenPixCount = screenWidth * screenHeight
)

type PixArray [screenPixCount * 4]uint8

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize glfw: ", err)
	}

	defer glfw.Terminate()

	window, err := glfw.CreateWindow(screenWidth, screenHeight, "Pixels", nil, nil)
	if err != nil {
		log.Fatalln("Failed to initialize window: ", err)
	}

	window.MakeContextCurrent()

	err = gl.Init()
	if err != nil {
		log.Fatalln("Failed to initialize gl: ", err)
	}

	var texture uint32
	gl.GenTextures(1, &texture)

	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	gl.BindImageTexture(0, texture, 0, false, 0, gl.WRITE_ONLY, gl.RGBA8)

	var framebuffer uint32
	gl.GenFramebuffers(1, &framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)

	pixels := PixArray{}

	for x := 0; x < screenWidth; x++ {
		for y := 0; y < screenHeight; y++ {
			setPixel(&pixels, x, y, uint8(float32(x)/screenWidth*255), uint8(float32(y)/screenHeight*255), 255)
		}
	}

	drawTriangle(&pixels, 40, 40, 20, 60, 60, 60)
	drawTriangle(&pixels, 20, 60, 60, 60, 40, 80)
	drawTriangle(&pixels, 80, 80, 60, 100, 100, 120)
	drawTriangle(&pixels, 0, 0, screenWidth, 0, screenWidth, screenHeight)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(screenWidth), int32(screenHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))

	for !window.ShouldClose() {
		windowWidth, windowHeight := window.GetSize()

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, screenWidth, screenHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))

		gl.BlitFramebuffer(0, 0, screenWidth, screenHeight, 0, 0, int32(windowWidth), int32(windowHeight), gl.COLOR_BUFFER_BIT, gl.NEAREST)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func setPixel(pixels *PixArray, x, y int, r, g, b uint8) {
	pixelI := (x + y*screenWidth) * 4
	pixels[pixelI] = r
	pixels[pixelI+1] = g
	pixels[pixelI+2] = b
	pixels[pixelI+3] = 255
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func drawTriangle(pixels *PixArray, x1, y1, x2, y2, x3, y3 int) {
	var topX, topY, midX, midY, btmX, btmY int

	if y1 < y2 {
		if y1 < y3 {
			if y2 < y3 {
				topX = x1
				topY = y1
				midX = x2
				midY = y2
				btmX = x3
				btmY = y3
			} else {
				topX = x1
				topY = y1
				midX = x3
				midY = y3
				btmX = x2
				btmY = y2
			}
		} else {
			topX = x3
			topY = y3
			midX = x1
			midY = y1
			btmX = x2
			btmY = y2
		}
	} else {
		if y2 < y3 {
			if y1 < y3 {
				topX = x2
				topY = y2
				midX = x1
				midY = y1
				btmX = x3
				btmY = y3
			} else {
				topX = x2
				topY = y2
				midX = x3
				midY = y3
				btmX = x1
				btmY = y1
			}
		} else {
			topX = x3
			topY = y3
			midX = x2
			midY = y2
			btmX = x1
			btmY = y1
		}
	}

	if midY == btmY {
		drawFlatBtmTriangle(pixels, topX, topY, midX, midY, btmX, btmY)
	} else if topY == midY {
		drawFlatTopTriangle(pixels, topX, topY, midX, midY, btmX, btmY)
	} else {
		newY := midY
		newX := int(float32(midY-topY)/float32(btmY-topY)*float32(btmX-topX)) + topX

		drawFlatBtmTriangle(pixels, topX, topY, midX, midY, newX, newY)
		drawFlatTopTriangle(pixels, midX, midY, newX, newY, btmX, btmY)
	}
}

func drawFlatBtmTriangle(pixels *PixArray, x1, y1, x2, y2, x3, y3 int) {
	minY := y1
	maxY := y3

	leftSlope := float32(min(x2, x3)-x1) / float32(maxY-minY)
	rightSlope := float32(max(x2, x3)-x1) / float32(maxY-minY)

	for y := minY; y < maxY; y++ {
		dy := y - minY
		minX := int(leftSlope*float32(dy)) + x1
		maxX := int(rightSlope*float32(dy)) + x1

		for x := minX; x < maxX; x++ {
			setPixel(pixels, x, y, 0, 0, 0)
		}
	}
}

func drawFlatTopTriangle(pixels *PixArray, x1, y1, x2, y2, x3, y3 int) {
	minY := y1
	maxY := y3

	leftSlope := float32(x3-min(x1, x2)) / float32(maxY-minY)
	rightSlope := float32(x3-max(x1, x2)) / float32(maxY-minY)

	for y := minY; y < maxY; y++ {
		dy := y - minY
		minX := int(leftSlope*float32(dy)) + min(x1, x2)
		maxX := int(rightSlope*float32(dy)) + max(x1, x2)

		for x := minX; x < maxX; x++ {
			setPixel(pixels, x, y, 0, 0, 0)
		}
	}
}
