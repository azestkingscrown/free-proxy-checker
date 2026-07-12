package models

// Proxy represents a single proxy and its validation statistics.
type Proxy struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Protocol  string `json:"protocol"`
	Latency   int64  `json:"latency"` // Latency in milliseconds
	Timestamp string `json:"timestamp"`
}
