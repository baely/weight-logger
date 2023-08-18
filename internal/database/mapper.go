package database

import (
	"github.com/baely/weightloss-tracker/internal/integrations/apple"
)

func ExportToDocuments(data apple.ExportData) []Document {
	documents := make(map[string]*Document)

	for _, metric := range data.Metrics {

		for _, metricData := range metric.Data {
			dateString := metricData.Date.Format("2006-01-02")

			if _, ok := documents[dateString]; !ok {
				documents[dateString] = &Document{Title: dateString}
			}

			switch metric.Name {
			case "active_energy":
				documents[dateString].ActiveEnergy = metricData.Quantity
			case "basal_energy_burned":
				documents[dateString].RestingEnergy = metricData.Quantity
			case "dietary_energy":
				documents[dateString].IntakeEnergy = metricData.Quantity
			case "weight_body_mass":
				documents[dateString].Weight = metricData.Quantity
			}
		}
	}

	flattenedDocuments := make([]Document, 0, len(documents))
	for _, document := range documents {
		flattenedDocuments = append(flattenedDocuments, *document)
	}

	return flattenedDocuments
}
