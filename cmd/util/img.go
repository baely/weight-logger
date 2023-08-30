package main

import (
	"fmt"
	"os"

	"github.com/baely/weightloss-tracker/internal/database"
	"github.com/baely/weightloss-tracker/internal/util/image"
)

func main() {
	doc := database.Document{
		Title:         "2023-05-05",
		ActiveEnergy:  3_000,
		RestingEnergy: 10_000.0,
		IntakeEnergy:  5_000.00001,
		Weight:        115.0,
	}

	img, err := image.Generate(doc)
	if err != nil {
		fmt.Println("error generating image:", err)
	}

	f, _ := os.Create("./img.jpg")

	_, err = f.Write(img)
	if err != nil {
		fmt.Println("error writing img:", err)
	}
}
