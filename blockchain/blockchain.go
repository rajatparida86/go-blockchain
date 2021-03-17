package blockchain

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "time"
)

type BlockChain struct {
    chain               []*Block
    currentTransactions []*Transaction
    nodes               map[string]bool
}

func (bc *BlockChain) Nodes() map[string]bool {
    return bc.nodes
}

func NewBlockChain() *BlockChain {
    bc := &BlockChain{
        chain:               make([]*Block, 0),
        currentTransactions: make([]*Transaction, 0),
        nodes:               make(map[string]bool, 0),
    }
    // Genesis block
    gen := bc.NewBlock()
    bc.ProofOfWork(gen)
    bc.AddBlock(gen)

    return bc
}

// Creates a new block
func (bc *BlockChain) NewBlock() *Block {
    var previousHash string

    if len(bc.chain) == 0 {
        previousHash = "1"
    } else {
        previousHash = bc.LastBlock().Hash
    }

    b := &Block{
        Index:        len(bc.chain) + 1,
        TimeStamp:    time.Now().Unix(),
        Transactions: bc.currentTransactions,
        Proof:        0,
        PreviousHash: previousHash,
    }
    return b
}

// Creates a new block and adds it to the block chain
func (bc *BlockChain) AddBlock(block *Block) {
    bc.chain = append(bc.chain, block)
}

// Adds new transaction to current transaction list of block chain and
// returns index for the next new block where the transaction will be added
func (bc *BlockChain) AddTransaction(tx *Transaction) int {
    bc.currentTransactions = append(bc.currentTransactions, tx)
    return bc.LastBlock().Index + 1
}

// Returns the latest block in block chain
func (bc *BlockChain) LastBlock() *Block {
    return bc.chain[len(bc.chain)-1]
}

// Returns the entire block chain
func (bc *BlockChain) Chain() []*Block {
    return bc.chain
}

// Registers a new node to the chain. This action is idempotent
func (bc *BlockChain) RegisterNode(address string) {
    bc.nodes[address] = true
}

// Proof of work algorithm
// - Finds the appropriate proof value
// - Sets the appropriate hash based on the appropriate proof
func (bc *BlockChain) ProofOfWork(block *Block) int {
    newProof := 0
    for {
        block.Proof = newProof
        if block.isProofValid() {
            block.Hash = block.GenerateHash()
            return newProof
        }
        newProof++
    }
}

// Check if the block chain is valid or not. Validity check:
// - A given Block should have a `proof` that results in the block hash to start with "0000"
// - Each block's `previousHash` should be equal to the hash of the previous block in the chain
func (bc *BlockChain) isChainValid() bool {
    valid := true
    for i, block := range bc.chain {
        // copy the block into a temp block without the hash
        // to check for proof of work
        tmp:= block.CopyWithoutHash()

        if valid = tmp.isProofValid(); !valid {
            break
        }
        if i > 0 {
            if valid = block.PreviousHash == bc.chain[i-1].Hash; !valid {
                //log.Printf("not previous hash for block - %v", i)
                log.Printf("index - %v \n propPrevHash - %v \n prev - %v", i, block.PreviousHash,
                    bc.chain[i-1].Hash)
                break
            }
        }
    }
    return valid
}

// Consensus algorithm - Replace current chain with the longest chain from the network
func (bc *BlockChain) ResolveConflicts() bool {
    var replaced = false
    var newChain = make([]*Block, 0)
    for node, _ := range bc.nodes {
        chain, err := bc.getChain(node)
        if err != nil {
            log.Printf("node with address - %s returned with err - %s", node, err)
            continue
        }
        remoteBlockChain := &BlockChain{
            chain: chain,
        }
        // Only take action if remote chain is longer than current node's chain
        if len(remoteBlockChain.chain) > len(bc.chain) && remoteBlockChain.isChainValid() {
            newChain = chain
            replaced = true
        }
    }
    if replaced {
        bc.chain = newChain
    }
    return replaced
}

func (bc *BlockChain) getChain(address string) ([]*Block, error) {
    chainResp := struct {
        Chain  []*Block `json:"chain"`
        Length int      `json:"length"`
    }{}

    url := fmt.Sprintf("http://%s/chain", address)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("node replied with %v", resp.StatusCode)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if err = json.Unmarshal(body, &chainResp); err != nil {
        return nil, err
    }

    return chainResp.Chain, nil
}
