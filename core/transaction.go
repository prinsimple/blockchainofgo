package core

type Transaction struct {
	ID      string
	From    string
	To      string
	Amount  int
	Payload []byte
}
