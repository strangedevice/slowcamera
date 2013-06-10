// Copyright (c) 2013 Antony Bowers <anton@strangedevice.co.uk>
// All rights reserved. 
// Use governed by the FreeBSD license. See LICENSE file.

package main

// #cgo CFLAGS:  -I/opt/vc/include
// #cgo LDFLAGS: -L/opt/vc/lib -lGLESv2
// #include "VG/openvg.h"
import "C"
import (
    "fmt"
    "reflect"
    "unsafe"
    "github.com/ajstarks/openvg"
    "code.google.com/p/go-opencv/trunk/opencv"
)

const (
    VG_CHANS = 4 // openVG raster is in RGBA format
)

var (
    sWidth, sHeight int // openVG screen dimensions
    iWidth, iHeight int // image dimensions
    out []C.VGubyte     // the openVG 8-bit RGBA interleaved raster 
)

// Initialise to given image size

func InitDisplay(iw, ih int) {
    sWidth, sHeight = openvg.Init()
    iWidth, iHeight = iw, ih
    out = make([]C.VGubyte, iw * ih * VG_CHANS)

    // clear screen to opaque black
    openvg.StartColor(sWidth, sHeight, "black", 1.0)
    openvg.End()

    // apply scale transform so images fill the screen
    ScaleImage(float64(sWidth) / float64(iWidth), 
        float64(sHeight) / float64(iHeight))

    fmt.Printf("Image display initialised to (%d, %d)\n", iw, ih)
}

func ScaleImage(xscale, yscale float64) {
    C.vgSeti(C.VG_MATRIX_MODE, C.VG_MATRIX_IMAGE_USER_TO_SURFACE)
    openvg.Scale(xscale, yscale)
}

// Low-level image drawing.
// Draws an openVG RGBA raster of size w*h at screen location (0,0).
// Assumes raster data is continuous.

func drawImage(w, h int, data unsafe.Pointer) {
    var stride C.VGint = C.VGint(w * 4)
    var format C.VGImageFormat = C.VG_sABGR_8888
    var quality C.VGbitfield =  C.VG_IMAGE_QUALITY_BETTER
    var img C.VGImage

    img = C.vgCreateImage(format, C.VGint(w), C.VGint(h), quality)
    C.vgImageSubData(img, data, stride, format,
        0, 0, C.VGint(w), C.VGint(h))

    // vgDrawImage() applies the scale transformation, 
    // vgSetPixels() does not
    C.vgDrawImage(img)
    C.vgDestroyImage(img)
}

// Hack to convert a C pointer to a Go byte slice

func toByteSlice(p unsafe.Pointer, length int) []byte {
    var bytes []byte

    header := (*reflect.SliceHeader)((unsafe.Pointer(&bytes)))
    header.Cap = length
    header.Len = length
    header.Data = uintptr(p)

    return bytes
}

// Convenience function to copy an entire image to the output

func DisplayImage(image *opencv.IplImage) {
    DisplaySubImage(0, 0, image.Width(), image.Height(), image)
}

// Copy a rectangular sub-image to the output,
// defined by top-left corner (itlx, itly) and size w * h

func DisplaySubImage(itlx, itly int, w, h int, image *opencv.IplImage) {
    DisplaySubImageShifted(itlx, itly, w, h, image, itlx, itly)
}

// Copy a rectangular sub-image to a different position in the output.
// Input has top-left corner (itlx, itly), output (otlx, otly).
// Both sub-images have size w * h and must lie within their respective
// containing images (no clipping).

func DisplaySubImageShifted(
        itlx, itly int, w, h int,
        image *opencv.IplImage,
        otlx, otly int) {

    channels := image.Channels()
    in := toByteSlice(image.ImageData(),
        image.Width() * image.Height() * channels)

    for iy, oy := itly, otly; iy < itly + h; iy, oy = iy + 1, oy + 1 {
        for ix, ox := itlx, otlx; ix < itlx + w; ix, ox = ix + 1, ox + 1 {
            // image origin is top-left
            imagep := (ix + (iHeight - 1 - iy) * iWidth) * channels
            // openVG origin is bottom-left
            outp := (ox + oy * iWidth) * VG_CHANS

            out[outp + 0] = C.VGubyte(in[imagep + 2]); // red
            out[outp + 1] = C.VGubyte(in[imagep + 1]); // green
            out[outp + 2] = C.VGubyte(in[imagep + 0]); // blue
            out[outp + 3] = 255 // alpha
        }
    }
}

// Sets the entire output image to opaque black
func Clear() {
    for i := range out {
        out[i] = 0
        if i % 4 == 3 {
            out[i] = 255 // alpha
        }
    }
}

// Displays the complete output image centered on the screen
func Show() {
    drawImage(iWidth, iHeight, unsafe.Pointer(&out[0]))
    openvg.End() // buffer swap
}

func FinishDisplay() {
    openvg.Finish()
    out = nil
    fmt.Printf("Finished with OpenVG\n")
}
