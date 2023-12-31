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
	err := c.loadFonts("Roboto-Regular.ttf", "CarterOne-Regular.ttf")
	if err != nil {
		return nil, err
	}

	c.SetClip(img.Bounds())
	c.SetDst(img)

	// Colors
	lightGrey := image.NewUniform(color.RGBA{R: 233, G: 233, B: 233, A: 255})
	red := image.NewUniform(color.RGBA{R: 201, G: 8, B: 79, A: 255})

	// Draw rectangles on the image
	rectangles := []image.Rectangle{
		{Min: image.Point{0, 0}, Max: image.Point{1080, 280}},
		{Min: image.Point{80, 400}, Max: image.Point{500, 650}},
		{Min: image.Point{80, 750}, Max: image.Point{500, 1000}},
		{Min: image.Point{580, 400}, Max: image.Point{1000, 650}},
		{Min: image.Point{580, 750}, Max: image.Point{1000, 1000}},
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
		{"CarterOne-Regular.ttf", 120, image.Black, fmt.Sprintf("Daily Update"), freetype.Pt(20, 125)},
		{"CarterOne-Regular.ttf", 120, image.Black, doc.Title, freetype.Pt(325, 250)},
		{"Roboto-Regular.ttf", 64, image.Black, fmt.Sprintf("Weight"), freetype.Pt(80, 390)},
		{"Roboto-Regular.ttf", 64, image.Black, fmt.Sprintf("Intake"), freetype.Pt(580, 390)},
		{"Roboto-Regular.ttf", 64, image.Black, fmt.Sprintf("Active Energy"), freetype.Pt(80, 740)},
		{"Roboto-Regular.ttf", 64, image.Black, fmt.Sprintf("Resting Energy"), freetype.Pt(580, 740)},
		// Units
		{"Roboto-Regular.ttf", 72, image.Black, "kg", freetype.Pt(400, 625)},
		{"Roboto-Regular.ttf", 72, image.Black, "kJ", freetype.Pt(900, 625)},
		{"Roboto-Regular.ttf", 72, image.Black, "kJ", freetype.Pt(400, 975)},
		{"Roboto-Regular.ttf", 72, image.Black, "kJ", freetype.Pt(900, 975)},
		// Qty
		{"CarterOne-Regular.ttf", 108, red, fmt.Sprintf("%.1f", doc.Weight), freetype.Pt(100, 525)},
		{"CarterOne-Regular.ttf", 108, red, fmt.Sprintf("%.0f", doc.IntakeEnergy), freetype.Pt(600, 525)},
		{"CarterOne-Regular.ttf", 108, red, fmt.Sprintf("%.0f", doc.ActiveEnergy), freetype.Pt(100, 875)},
		{"CarterOne-Regular.ttf", 108, red, fmt.Sprintf("%.0f", doc.RestingEnergy), freetype.Pt(600, 875)},
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
