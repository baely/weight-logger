package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"

	"github.com/baely/weightloss-tracker/internal/database"
	"github.com/baely/weightloss-tracker/internal/integrations/gcs"
	"github.com/baely/weightloss-tracker/internal/util"
)

const (
	FilenameFormat = "weightlog/%s.jpg"
)

// Context wraps the freetype context and provides utility methods for font operations
type context struct {
	*freetype.Context
	fonts map[string]*truetype.Font
}

// NewContext initializes a new freetype context
func newContext() *context {
	return &context{
		Context: freetype.NewContext(),
	}
}

// LoadFonts loads the specified fonts into the context
func (c *context) loadFonts(fonts ...string) error {
	fontMap := make(map[string]*truetype.Font)

	for _, fontName := range fonts {
		fontFile, err := gcs.ReadFile(util.StaticResourceBucket, fontName)
		if err != nil {
			return err
		}
		b, err := io.ReadAll(fontFile)
		if err != nil {
			return err
		}
		f, err := truetype.Parse(b)
		if err != nil {
			return err
		}
		fontMap[fontName] = f
	}

	c.fonts = fontMap
	return nil
}

// SetFont sets the font for the context
func (c *context) setFont(font string) {
	if f, ok := c.fonts[font]; ok {
		c.SetFont(f)
	}
}

// WriteString draws a string onto the image using the specified parameters
func (c *context) writeString(font string, size float64, src image.Image, text string, point fixed.Point26_6) error {
	c.setFont(font)
	c.SetFontSize(size)
	c.SetSrc(src)
	_, err := c.DrawString(text, point)
	return err
}

// Generate creates and returns an image based on the provided document data
func Generate(doc database.Document) ([]byte, error) {
	width, height := 1080, 1080
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Set background color
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// Initialize freetype context and load fonts
	c := newContext()
	err := c.loadFonts("Roboto-Regular.ttf", "AbrilFatface-Regular.ttf", "AzeretMono-Bold.ttf", "B612Mono-Bold.ttf", "AzeretMono-Bold.ttf")
	if err != nil {
		return nil, err
	}

	c.SetClip(img.Bounds())
	c.SetDst(img)

	// Colors
	lightGrey := image.NewUniform(color.RGBA{R: 225, G: 225, B: 225, A: 255})
	red := image.NewUniform(color.RGBA{R: 201, G: 8, B: 79, A: 255})

	// Draw rectangles on the image
	rectangles := []image.Rectangle{
		{Min: image.Point{0, 0}, Max: image.Point{1080, 180}},
		{Min: image.Point{125, 400}, Max: image.Point{500, 550}},
		{Min: image.Point{125, 750}, Max: image.Point{500, 900}},
		{Min: image.Point{580, 400}, Max: image.Point{955, 550}},
		{Min: image.Point{580, 750}, Max: image.Point{955, 900}},
	}
	for _, r := range rectangles {
		draw.Draw(img, r, lightGrey, image.Pt(0, 0), draw.Src)
	}

	// Draw text onto the image
	texts := []struct {
		font  string
		size  float64
		src   image.Image
		text  string
		point fixed.Point26_6
	}{
		// Title fonts
		{"AbrilFatface-Regular.ttf", 64, image.Black, fmt.Sprintf("Daily Update     %s", doc.Title), freetype.Pt(150, 150)},
		{"Roboto-Regular.ttf", 32, image.Black, fmt.Sprintf("Weight"), freetype.Pt(145, 390)},
		{"Roboto-Regular.ttf", 32, image.Black, fmt.Sprintf("Intake"), freetype.Pt(600, 390)},
		{"Roboto-Regular.ttf", 32, image.Black, fmt.Sprintf("Active Energy"), freetype.Pt(145, 740)},
		{"Roboto-Regular.ttf", 32, image.Black, fmt.Sprintf("Resting Energy"), freetype.Pt(600, 740)},
		// Units
		{"Roboto-Regular.ttf", 58, image.Black, "kg", freetype.Pt(400, 500)},
		{"Roboto-Regular.ttf", 58, image.Black, "kJ", freetype.Pt(855, 500)},
		{"Roboto-Regular.ttf", 58, image.Black, "kJ", freetype.Pt(400, 850)},
		{"Roboto-Regular.ttf", 58, image.Black, "kJ", freetype.Pt(855, 850)},
		// Qty
		{"AbrilFatface-Regular.ttf", 72, red, fmt.Sprintf("%.1f", doc.Weight), freetype.Pt(165, 500)},
		{"AbrilFatface-Regular.ttf", 72, red, fmt.Sprintf("%.0f", doc.IntakeEnergy), freetype.Pt(615, 500)},
		{"AbrilFatface-Regular.ttf", 72, red, fmt.Sprintf("%.0f", doc.ActiveEnergy), freetype.Pt(165, 850)},
		{"AbrilFatface-Regular.ttf", 72, red, fmt.Sprintf("%.0f", doc.RestingEnergy), freetype.Pt(615, 850)},
	}
	for _, text := range texts {
		err = c.writeString(text.font, text.size, text.src, text.text, text.point)
		if err != nil {
			return nil, err
		}
	}

	// Encode the image to JPEG format
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
