package bursary

type Rule struct {
	Commission    float64 `json:"commission"`
	Share         float64 `json:"share"`
	ReturnedShare float64 `json:"returned_share"`
}

var DefaultRule = Rule{
	Commission:    0.0, // No commissions for returning
	Share:         0.0, // used to give share to current member ()
	ReturnedShare: 0.0, // used to return share for upstream (upstream's share >= share + returned share)
}
