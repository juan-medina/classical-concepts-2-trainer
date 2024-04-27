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

package shapes

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteImage    = ebiten.NewImage(3, 3)
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	b := whiteImage.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	for i := range pix {
		pix[i] = 0xFF
	}
	whiteImage.WritePixels(pix)
}

func getXYFromCenterWithAngleRadius(centerX, centerY, angle, radius float32) (float32, float32) {
	return centerX + radius*float32(math.Cos(float64(angle))), centerY + radius*float32(math.Sin(float64(angle)))
}

func DrawPolygon(dst *ebiten.Image, centerX float32, centerY float32, radius float32, sides int, rotation float32, color color.Color) {
	centerAngle := float32(rotation * math.Pi / 180.0)
	angleStep := 360.0 / float32(sides) * math.Pi / 180.0

	var path = vector.Path{}
	for i := 0; i < sides; i++ {
		path.LineTo(getXYFromCenterWithAngleRadius(centerX, centerY, centerAngle, radius))
		centerAngle += angleStep
	}

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	r, g, b, a := color.RGBA()
	cr, cg, cb, ca := float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff

	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = cr
		vs[i].ColorG = cg
		vs[i].ColorB = cb
		vs[i].ColorA = ca
	}

	dst.DrawTriangles(vs, is, whiteSubImage, &ebiten.DrawTrianglesOptions{})
}
