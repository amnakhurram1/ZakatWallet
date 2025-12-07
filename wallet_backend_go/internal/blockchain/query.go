package blockchain

// query.go adds helper methods to read data from the blockchain,
// which we use for the block explorer and wallet history APIs.

import (
    "bytes"
    "encoding/hex"
    "errors"
)

// BlockSummary is a lightweight view of a block for list endpoints.
type BlockSummary struct {
    Index     int    `json:"index"`
    Timestamp int64  `json:"timestamp"`
    Hash      string `json:"hash"`
    PrevHash  string `json:"prev_hash"`
    TxCount   int    `json:"tx_count"`
}

// ListBlocks returns basic info about all blocks in the chain.
func (bc *Blockchain) ListBlocks() []BlockSummary {
    summaries := make([]BlockSummary, 0, len(bc.Blocks))
    for i, b := range bc.Blocks {
        summaries = append(summaries, BlockSummary{
            Index:     i,
            Timestamp: b.Timestamp,
            Hash:      hex.EncodeToString(b.Hash),
            PrevHash:  hex.EncodeToString(b.PrevHash),
            TxCount:   len(b.Transactions),
        })
    }
    return summaries
}

// GetBlockByIndex returns a block by its index in the slice.
func (bc *Blockchain) GetBlockByIndex(idx int) (*Block, bool) {
    if idx < 0 || idx >= len(bc.Blocks) {
        return nil, false
    }
    return bc.Blocks[idx], true
}

// GetTransactionsForAddress returns all transactions that have
// at least one output paying to the given wallet address.
func (bc *Blockchain) GetTransactionsForAddress(address string) ([]*Transaction, error) {
    if !ValidateAddress(address) {
        return nil, errors.New("invalid address")
    }

    pubKeyHash, err := hex.DecodeString(address)
    if err != nil {
        return nil, errors.New("invalid address encoding")
    }

    var txs []*Transaction
    for _, b := range bc.Blocks {
        for _, tx := range b.Transactions {
            // Check outputs only (receiving side). We can extend later
            // to also detect "sent" transactions.
            for _, out := range tx.Vout {
                if bytes.Equal(out.PubKeyHash, pubKeyHash) {
                    txs = append(txs, tx)
                    break
                }
            }
        }
    }
    return txs, nil
}
