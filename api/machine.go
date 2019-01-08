package api

type Machine struct {
	ID        string `json:"id"`
	CPU       int    `json:"cpu"`
	Memory    int    `json:"memory"`
	NodeLabel string `json:"nodeLabel"`
	GPUCount  int    `json:"gpuCount,omitempty"`
	GPUType   string `json:"gpuType,omitempty"`
	GPULabel  string `json:"gpuLabel,omitempty"`
	Cost      int64  `json:"cost"`
	IsActive  bool   `json:"isActive"`
}
