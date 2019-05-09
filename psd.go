package main

import (
	"fmt"
	"github.com/oov/psd"
	"image"
	"image/draw"
	"image/png"
	"os"
)

/**
 * processAndPrintLayer
 * Processing the top-level layers, getting all the images,
 * drawing them over one another and then saving to .png
 */
func processAndPrintLayer (layer psd.Layer) error {
	// go through all sub layers and pull out all the images
	var childImages []image.Image
	for _, subLayer := range layer.Layer {
		processSubLayer(subLayer, &childImages)
	}

	if len(childImages) == 0 {
		return nil
	}

	// draw all the images together using the bounds from the mask
	newImage := image.NewRGBA(layer.Mask.Rect)
	for _, img := range childImages {
		draw.Draw(newImage, img.Bounds(), img, img.Bounds().Min, draw.Over)
	}

	// finally, create a .png file and print it out
	out, err := os.Create(fmt.Sprintf("./output/%s.png", layer.Name))
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, newImage)

}

/**
 * processSubLayer
 * Go through and process subLayers. If it has an image, add it to childImages,
 * and if it has sub-sub-layers, add it to childImages recursively.
 */
func processSubLayer (subLayer psd.Layer, childImages *[]image.Image) {
	// add the sublayer's image to the childImages array
	if subLayer.HasImage() {
		*childImages = append(*childImages, subLayer.Picker)
	}

	// recursively add the sub-sublayer's images to the childImages array
	if len(subLayer.Layer) > 0 {
		for _, subSubLayer := range subLayer.Layer {
			processSubLayer(subSubLayer, childImages)
		}
	}
}

func main() {
	// get args for input
	args := os.Args[1:]

	if len(args) != 1 {
		fmt.Printf("Usage: go run psd.go /path/to/file.psd")
		os.Exit(1)
	}
	// open the file
	file, err := os.Open(args[0])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// do PSD decode magic to get our root PSD object
	root, _, err := psd.Decode(file, &psd.DecodeOptions{SkipMergedImage: true})
	if err != nil {
		panic(err)
	}

	// get rid of all non-visible layers
	var visibleLayers []psd.Layer
	for _, layer := range root.Layer {
		if layer.Visible() {
			visibleLayers = append(visibleLayers, layer)
		}
	}
	// finally process all remaining visible layers
	for _, layer := range visibleLayers {
		processAndPrintLayer(layer)
	}
}
