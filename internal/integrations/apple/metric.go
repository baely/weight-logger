package apple

import (
	"strings"
	"time"
)

type Export struct {
	Data ExportData `json:"data"`
}

type ExportData struct {
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	Name string       `json:"name"`
	Unit string       `json:"unit"`
	Data []MetricData `json:"data"`
}

type MetricData struct {
	Source   string         `json:"source"`
	Date     MetricDataDate `json:"date"`
	Quantity float64        `json:"qty"`
}

type MetricDataDate struct {
	time.Time
}

func (m *MetricDataDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		m.Time = time.Time{}
		return nil
	}

	var err error
	m.Time, err = time.Parse("2006-01-02 15:04:05 -0700", s)
	return err
}
