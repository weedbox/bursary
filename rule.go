package main

type Rule struct {
	Commission float64 `json:"commission"`
	Share      float64 `json:"share"`
}

var DefaultRule = Rule{
	Commission: 0,   // No commissions for returning
	Share:      1.0, // Take entire rewards
}
