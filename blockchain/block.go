package blockchain

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "strings"
)

type Block struct {
    Index        int            `json:"index"`
    TimeStamp    int64          `json:"time_stamp"`
    Transactions []*Transaction `json:"transactions"`
    Proof        int            `json:"proof"`
    Hash         string         `json:"hash"`
    PreviousHash string         `json:"previous_hash"`
}

func (b *Block) NewTransaction() uint {
    return 0
}

func (b *Block) GenerateHash() string {
    blockBytes, _ := json.Marshal(b)
    hash := sha256.Sum256(blockBytes)
    return hex.EncodeToString(hash[:])
}

// validate proof
func (b *Block) isProofValid() bool {
    return strings.HasPrefix(b.GenerateHash(), "0000")
}

func (b *Block) CopyWithoutHash() *Block  {
    return &Block{
        Index:        b.Index,
        TimeStamp:    b.TimeStamp,
        Transactions: b.Transactions,
        Proof:        b.Proof,
        Hash:         "",
        PreviousHash: b.PreviousHash,
    }
}
