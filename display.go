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
    fmt.Printf("Image display initialised to (%d, %d)\n", iw, ih)
}

// Low-level image drawing.
// Draws an openVG RGBA raster of size w*h at screen location (x,y).
// Assumes raster data is continuous.

func drawImage(x, y int, w, h int, data unsafe.Pointer) {
    var stride C.VGint = C.VGint(w * 4)
    var format C.VGImageFormat = C.VG_sABGR_8888
    var quality C.VGbitfield =  C.VG_IMAGE_QUALITY_BETTER
    var img C.VGImage

    img = C.vgCreateImage(format, C.VGint(w), C.VGint(h), quality)
    C.vgImageSubData(img, data, stride, format,
        0, 0, C.VGint(w), C.VGint(h))
    C.vgSetPixels(C.VGint(x), C.VGint(y), img, 0, 0, C.VGint(w), C.VGint(h))
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
// defined by top-left corner (tlx, tly) and size w x h

func DisplaySubImage(tlx, tly int, w, h int, image *opencv.IplImage) {
    DisplaySubImageShifted(tlx, tly, w, h, image, tlx, tly)
}

func DisplaySubImageShifted(
        itlx, itly int, w, h int,
        image *opencv.IplImage,
        otlx, otly int) {

    channels := image.Channels()
    in := toByteSlice(image.ImageData(),
        image.Width() * image.Height() * channels)

    for iy, oy := itly, otly; iy < itly + h; iy++ {
        for ix, ox := itlx, otlx; ix < itlx + w; ix++ {
            // image origin is top-left
            imagep := (ix + (iHeight - 1 - iy) * iWidth) * channels
            // openVG origin is bottom-left
            outp := (ox + oy * iWidth) * VG_CHANS

            out[outp + 0] = C.VGubyte(in[imagep + 2]); // red
            out[outp + 1] = C.VGubyte(in[imagep + 1]); // green
            out[outp + 2] = C.VGubyte(in[imagep + 0]); // blue
            out[outp + 3] = 255 // alpha

            ox++
        }

        oy++
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
    openvg.StartColor(sWidth, sHeight, "black", 1.0)
}

// Displays the complete output image centered on the screen
func Show() {
    xo := (sWidth - iWidth) / 2
    yo := (sHeight - iHeight) / 2
    drawImage(xo, yo, iWidth, iHeight, unsafe.Pointer(&out[0]))
    openvg.End() // buffer swap
}

func FinishDisplay() {
    openvg.Finish()
    out = nil
    fmt.Printf("Finished with OpenVG\n")
}
