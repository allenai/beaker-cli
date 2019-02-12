package api

import (
	"time"

	"github.com/shopspring/decimal"
)

// TODO: Document
type UsageReport struct {
	Start        time.Time     `json:"start"`
	End          time.Time     `json:"end"`
	Currency     string        `json:"currency"`
	EntityType   string        `json:"entityType"`
	Interval     string        `json:"interval"`
	IntervalKeys []time.Time   `json:"intervalKeys"`
	Items        []EntityUsage `json:"items"`
}

// TODO: Document
type EntityUsage struct {
	Entity    User                        `json:"entity"`
	Totals    UsageInterval               `json:"totals"`
	Intervals map[time.Time]UsageInterval `json:"intervals"`
}

// TODO: Document
type UsageInterval struct {
	ExperimentCount int             `json:"experimentCount"`
	Cost            decimal.Decimal `json:"cost"`
	Duration        int64           `json:"duration"`
	GPUSeconds      int64           `json:"gpuSeconds"`
}
