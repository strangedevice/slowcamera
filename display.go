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
    openvg.StartColor(sWidth, sHeight, "black")
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

func DisplayIplImage(x, y int, image *opencv.IplImage) {
    w, h := image.Width(), image.Height()
    fmt.Printf("Displaying image of size (%d, %d) at (%d, %d)\n", 
        w, h, x, y)

    in := toByteSlice(image.ImageData(), w * h * image.Channels())
    n := 0;

    for yp := 0; yp < h; yp++ {
        for xp := 0; xp < w; xp++ {
            offset := (xp + (h - 1 - yp) * w ) * image.Channels()

            out[n] = C.VGubyte(in[offset+2]); // red
            n++
            out[n] = C.VGubyte(in[offset+1]); // green
            n++
            out[n] = C.VGubyte(in[offset+0]); // blue
            n++
            out[n] = 255 // alpha
            n++
        }
    }
}

// Convenience function to copy an entire image to the output

func DisplayImage(image *opencv.IplImage) {
    DisplaySubImage(0, 0, image.Width(), image.Height(), image)
}

// Copy a rectangular sub-image to the output,
// defined by top-left corner (tlx, tly) and size w x h

func DisplaySubImage(tlx, tly int, w, h int, image *opencv.IplImage) {
    // fmt.Printf("Displaying sub-image of size (%d, %d) at (%d, %d)\n", 
    //     w, h, tlx, tly)

    channels := image.Channels()
    in := toByteSlice(image.ImageData(), 
        image.Width() * image.Height() * channels)

    for yp := tly; yp < tly + h; yp++ {
        for xp := tlx; xp < tlx + w; xp++ {
            // image origin is top-left
            imagep := (xp + (iHeight - 1 - yp) * iWidth) * channels
            // openVG origin is bottom-left
            outp := (xp + yp * iWidth) * VG_CHANS

            out[outp + 0] = C.VGubyte(in[imagep + 2]); // red
            out[outp + 1] = C.VGubyte(in[imagep + 1]); // green
            out[outp + 2] = C.VGubyte(in[imagep + 0]); // blue
            out[outp + 3] = 255 // alpha
        }
    }
}

// Sets the entire output image to transparent black
func Clear() {
    for i := range out {
        out[i] = 0
    }
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
