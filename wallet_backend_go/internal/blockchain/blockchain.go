package blockchain

// blockchain.go implements a minimal blockchain with proof‑of‑work and
// UTXO support. This in‑memory implementation demonstrates block
// addition, transaction lookup, UTXO scanning and simple PoW mining.
// For persistence you should store blocks in a database such as
// Supabase or another PostgreSQL backend via the db package.

import (
    "crypto/ecdsa"
    "encoding/hex"
    "fmt"
)

// Blockchain represents a chain of blocks. Blocks are kept in a slice
// for simplicity. In production you'd want a database indexed by
// block hashes, heights, etc. The Genesis block is at index 0.
type Blockchain struct {
    Blocks []*Block
}

// NewBlockchain creates a blockchain with a genesis block paying a
// reward to the provided address. It returns a pointer to the
// blockchain. Persisting the chain is left to the caller.
func NewBlockchain(address string) *Blockchain {
    coinbase := NewCoinbaseTx(address, "Genesis Block")
    genesis := NewBlock([]*Transaction{coinbase}, []byte{})
    bc := &Blockchain{Blocks: []*Block{genesis}}
    return bc
}

// AddBlock mines a new block containing the provided transactions.
// Proof‑of‑work is performed automatically via the NewBlock call.
// The new block is appended to the chain and returned. In a real
// system you'd also validate transactions and persist the block.
func (bc *Blockchain) AddBlock(txs []*Transaction) *Block {
    prevHash := bc.Blocks[len(bc.Blocks)-1].Hash
    newBlock := NewBlock(txs, prevHash)
    bc.Blocks = append(bc.Blocks, newBlock)
    return newBlock
}

// FindTransaction searches for a transaction by its ID and returns
// it. An error is returned if the transaction is not found in the
// chain. This method scans the blockchain linearly.
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
    for _, block := range bc.Blocks {
        for _, tx := range block.Transactions {
            if hex.EncodeToString(tx.ID) == hex.EncodeToString(ID) {
                return *tx, nil
            }
        }
    }
    return Transaction{}, fmt.Errorf("transaction not found")
}

// FindUTXO scans the entire blockchain and returns a map of
// unspent transaction outputs. If pubKeyHash is nil, all UTXOs are
// returned; otherwise only outputs matching the provided pubKeyHash
// are collected. The returned map is keyed by transaction ID hex
// strings with values being slices of TxOutput.
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) map[string][]TxOutput {
    spentTXOs := make(map[string][]int)
    UTXOs := make(map[string][]TxOutput)

    for _, block := range bc.Blocks {
        for _, tx := range block.Transactions {
            txIDStr := hex.EncodeToString(tx.ID)
            // iterate outputs
            for outIdx, out := range tx.Vout {
                // check if output is spent
                if spent, ok := spentTXOs[txIDStr]; ok {
                    skip := false
                    for _, spentOutIdx := range spent {
                        if spentOutIdx == outIdx {
                            skip = true
                            break
                        }
                    }
                    if skip {
                        continue
                    }
                }
                if pubKeyHash == nil || string(out.PubKeyHash) == string(pubKeyHash) {
                    UTXOs[txIDStr] = append(UTXOs[txIDStr], out)
                }
            }
            // record spent outputs
            if !tx.IsCoinbase() {
                for _, in := range tx.Vin {
                    if pubKeyHash == nil || true {
                        inTxID := hex.EncodeToString(in.Txid)
                        spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
                    }
                }
            }
        }
    }
    return UTXOs
}

// SignTransaction finds the referenced previous transactions and
// delegates signing to the transaction itself. It panics if any
// referenced transaction cannot be found. The caller is responsible
// for ensuring that the private key corresponds to the spender's
// public key.
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
    prevTXs := make(map[string]Transaction)
    for _, vin := range tx.Vin {
        prevTx, err := bc.FindTransaction(vin.Txid)
        if err != nil {
            return err
        }
        prevTXs[fmt.Sprintf("%x", vin.Txid)] = prevTx
    }
    return tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies the signatures on the transaction inputs.
// It looks up the previous transactions referenced by the inputs and
// passes them to the Verify method. Returns true if all signatures
// are valid. Coinbase transactions are always valid.
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
    if tx.IsCoinbase() {
        return true
    }
    prevTXs := make(map[string]Transaction)
    for _, vin := range tx.Vin {
        prevTx, err := bc.FindTransaction(vin.Txid)
        if err != nil {
            return false
        }
        prevTXs[fmt.Sprintf("%x", vin.Txid)] = prevTx
    }
    return tx.Verify(prevTXs)
}