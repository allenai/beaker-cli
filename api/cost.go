package api

// Bill describes a cost
type Bill struct {
	// A string representation of the dollar amount.
	// e.g. "0.0627775092"
	Value string `json:"value"`

	// Whether the value/amount is the final.
	// e.g. This would be False if a task or an experiment is still in progress.
	IsFinal bool `json:"isFinal"`
}
