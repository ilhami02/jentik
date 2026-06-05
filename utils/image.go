package utils

import (
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
)

// CompressAndSaveImage decodes an image from src, compresses it as JPEG with quality, and saves it to destPath.
func CompressAndSaveImage(src io.Reader, destPath string, quality int) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
}
