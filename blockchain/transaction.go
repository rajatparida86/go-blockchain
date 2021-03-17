package blockchain

type Transaction struct {
    Sender   string `json:"sender"`
    Receiver string `json:"receiver"`
    Amount   float64 `json:"amount"`
}
