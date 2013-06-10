// Copyright (c) 2013 Antony Bowers <anton@strangedevice.co.uk>
// All rights reserved. 
// Use governed by the FreeBSD license. See LICENSE file.

package main

import (
    "fmt"
    "os"
    "bufio"
    "code.google.com/p/go-opencv/trunk/opencv"
)

const (
    IMAGE_WIDTH = 320
    IMAGE_HEIGHT = 240
    STRIPE_WIDTH = 1
)

func reportParameters(capture *opencv.Capture) {
    // Hue and Exposure are not supported in v4l
    bri := capture.GetProperty(opencv.CV_CAP_PROP_BRIGHTNESS)
    con := capture.GetProperty(opencv.CV_CAP_PROP_CONTRAST)
    sat := capture.GetProperty(opencv.CV_CAP_PROP_SATURATION)
    gai := capture.GetProperty(opencv.CV_CAP_PROP_GAIN)

    fmt.Printf("Brightness: %f\n", bri)
    fmt.Printf("Contrast: %f\n", con)
    fmt.Printf("Saturation: %f\n", sat)
    fmt.Printf("Gain: %f\n", gai)
}

func main() {

    InitDisplay(IMAGE_WIDTH, IMAGE_HEIGHT)
    defer FinishDisplay()

    capture := opencv.NewCameraCapture(0)
    if capture == nil {
        panic("Failed to find a camera")
    }
    defer capture.Release()

    capture.SetProperty(opencv.CV_CAP_PROP_FRAME_WIDTH, float64(IMAGE_WIDTH))
    capture.SetProperty(opencv.CV_CAP_PROP_FRAME_HEIGHT, float64(IMAGE_HEIGHT))

    capture.SetProperty(opencv.CV_CAP_PROP_BRIGHTNESS, 0.75)
    capture.SetProperty(opencv.CV_CAP_PROP_CONTRAST, 0.18)
    capture.SetProperty(opencv.CV_CAP_PROP_SATURATION, 0.2)
    capture.SetProperty(opencv.CV_CAP_PROP_GAIN, 0.05)

    reportParameters(capture)

    for {
        // simulate a button
        fmt.Printf("Press RETURN to start capturing...")
        bufio.NewReader(os.Stdin).ReadBytes('\n')

        Clear() // blank the output image

        for stripe := 0; stripe < IMAGE_WIDTH; stripe += STRIPE_WIDTH {
            image := capture.QueryFrame() // returns an IplImage
            if image == nil {
                panic("Failed to capture an image")
            }
            DisplaySubImage(stripe, 0, STRIPE_WIDTH, IMAGE_HEIGHT, image)

            // Crazy version - takes stripes from the centre of the image only
//            DisplaySubImageShifted(IMAGE_WIDTH / 2, 0, 2, IMAGE_HEIGHT, image,
//                stripe, 0)
            Show()
        }
    }
}
