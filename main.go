package main

import (
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    bc "github.com/rajatparida86/go-blockchain/blockchain"
    "log"
    "net/http"
    "os"
)

func main() {
    app := AppInit()

    http.HandleFunc("/transactions/new", app.newTransaction)
    http.HandleFunc("/mine", app.mine)
    http.HandleFunc("/chain", app.chain)
    http.HandleFunc("/nodes/register", app.registerNode)
    http.HandleFunc("/nodes/resolve", app.resolveConflicts)

    port := os.Getenv("blockPort")

    log.Println("listening on port - ", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

type app struct {
    blockChain *bc.BlockChain
    id         string
}

func AppInit() *app {
    nodeUUID := uuid.New().String()
    log.Println("initialising new node with ID: ", nodeUUID)
    return &app{
        blockChain: bc.NewBlockChain(),
        id:         nodeUUID,
    }
}

func (a *app) newTransaction(w http.ResponseWriter, r *http.Request) {
    var tx bc.Transaction
    dec := json.NewDecoder(r.Body)
    if err := dec.Decode(&tx); err != nil {
        a.writeSimpleResponse(w, http.StatusBadRequest)
        return
    }

    index := a.blockChain.AddTransaction(&tx)
    resp := struct {
        Message string
    }{
        Message: fmt.Sprintf("Transaction will be added to block with index %v", index),
    }
    a.writeResponse(w, http.StatusCreated, resp)
}

// Create proof of work
// Reward the miner with 1 coin
// Forge the new block by adding it to the chain
func (a *app) mine(w http.ResponseWriter, r *http.Request) {
    a.blockChain.AddTransaction(&bc.Transaction{
        Sender:   "0",
        Receiver: a.id,
        Amount:   1,
    })
    block := a.blockChain.NewBlock()

    a.blockChain.ProofOfWork(block)

    a.blockChain.AddBlock(block)

    resp := struct {
        Message string    `json:"message"`
        Block   *bc.Block `json:"block"`
    }{
        Message: "New block created and added to chain",
        Block:   block,
    }
    a.writeResponse(w, http.StatusOK, resp)
}

func (a *app) chain(w http.ResponseWriter, r *http.Request) {
    resp := struct {
        Chain  []*bc.Block `json:"chain"`
        Length int         `json:"length"`
    }{
        a.blockChain.Chain(),
        len(a.blockChain.Chain()),
    }
    a.writeResponse(w, http.StatusOK, resp)
}

func (a *app) registerNode(w http.ResponseWriter, r *http.Request) {
    req := struct {
        Nodes []string `json:"nodes"`
    }{}
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&req); err != nil {
        a.writeSimpleResponse(w, http.StatusBadRequest)
    }
    for _, node := range req.Nodes {
        a.blockChain.RegisterNode(node)
    }
    nodes := make([]string, 0)
    for node, _ := range a.blockChain.Nodes() {
        nodes = append(nodes, node)
    }

    resp := struct {
        Message string   `json:"message"`
        Nodes   []string `json:"nodes"`
    }{
        "New nodes added",
        nodes,
    }

    a.writeResponse(w, http.StatusCreated, resp)
}

func (a *app) resolveConflicts(w http.ResponseWriter, r *http.Request) {
    replaced := a.blockChain.ResolveConflicts()
    resp := struct {
        Message string      `json:"message"`
        Chain   []*bc.Block `json:"chain"`
    }{}
    if replaced {
        resp.Message = "Chain is replaced"
    } else {
        resp.Message = "Our chain is master"
    }
    resp.Chain = a.blockChain.Chain()
    a.writeResponse(w, http.StatusCreated, resp)
}

func (a *app) writeResponse(w http.ResponseWriter, status int, data interface{}) {
    resp, err := json.Marshal(data)
    if err != nil {
        a.writeSimpleResponse(w, http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _, _ = w.Write(resp)
}

func (a *app) writeSimpleResponse(w http.ResponseWriter, status int) {
    w.WriteHeader(status)
    _, _ = w.Write([]byte(http.StatusText(status)))
}
