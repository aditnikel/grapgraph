package domain

type Transfer struct {
	Tx     string
	From   string
	To     string
	Amount float64
	Ts     int64
}
