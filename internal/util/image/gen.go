package image

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"log"

	"github.com/golang/freetype"

	"github.com/baely/weightloss-tracker/internal/database"
	"github.com/baely/weightloss-tracker/internal/integrations/gcs"
	"github.com/baely/weightloss-tracker/internal/util"
)

const (
	FilenameFormat = "weightlog/%s.jpg"
)

func Generate(doc database.Document) ([]byte, error) {
	width, height := 1080, 1080
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the image with the color
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// Initialize freetype context
	c := freetype.NewContext()
	//c.SetDPI(72)
	file, err := gcs.ReadFile(util.StaticResourceBucket, "Roboto-Regular.ttf")
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	f, err := freetype.ParseFont(b)
	if err != nil {
		return nil, err
	}
	c.SetFont(f)
	//c.SetFont(freetype.PublicFont("luxi.ttf", freetype.LuxiSans))
	c.SetFontSize(24)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black) // Black text

	_, err = c.DrawString(fmt.Sprintf("title: %s", doc.Title), freetype.Pt(200, 200))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_, err = c.DrawString(fmt.Sprintf("active energy: %f", doc.ActiveEnergy), freetype.Pt(200, 300))
	_, err = c.DrawString(fmt.Sprintf("passive energy: %f", doc.RestingEnergy), freetype.Pt(200, 400))
	_, err = c.DrawString(fmt.Sprintf("consumed energy: %f", doc.IntakeEnergy), freetype.Pt(200, 500))
	_, err = c.DrawString(fmt.Sprintf("weight: %f", doc.Weight), freetype.Pt(200, 600))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var buf bytes.Buffer

	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
