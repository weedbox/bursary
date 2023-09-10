package bursary

type Rule struct {
	Commission float64 `json:"commission"`
	Share      float64 `json:"share"`
}

var DefaultRule = Rule{
	Commission: 0.0, // No commissions for returning
	Share:      0.0, // just pass trough
}
