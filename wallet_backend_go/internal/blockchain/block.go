package blockchain

// block.go defines the Block type and implements basic block
// construction and hashing logic. A block contains a slice of
// transactions, a previous block hash, its own hash and the nonce
// produced by the proof‑of‑work.

import (
    "bytes"
    "crypto/sha256"
    "time"
)

// Block represents a single block in the chain. Each block holds
// references to its parent via PrevHash, a slice of transactions,
// its own computed Hash and the Nonce discovered during mining.
type Block struct {
    Timestamp    int64
    Transactions []*Transaction
    PrevHash     []byte
    Hash         []byte
    Nonce        int
}

// NewBlock creates and returns a new block containing the provided
// transactions and the given previous hash. A proof‑of‑work is run
// internally to find a valid nonce and produce the block's hash.
func NewBlock(transactions []*Transaction, prevHash []byte) *Block {
    block := &Block{Timestamp: time.Now().Unix(), Transactions: transactions, PrevHash: prevHash, Hash: []byte{}, Nonce: 0}
    pow := NewProofOfWork(block)
    nonce, hash := pow.Run()
    block.Hash = hash[:]
    block.Nonce = nonce
    return block
}

// HashTransactions computes a single SHA‑256 hash over all
// transaction IDs in the block. This is a simplified Merkle tree
// implementation suitable for small blocks. The result is used as
// part of the proof‑of‑work input.
func (b *Block) HashTransactions() []byte {
    var txHashes [][]byte
    for _, tx := range b.Transactions {
        txHashes = append(txHashes, tx.ID)
    }
    data := bytes.Join(txHashes, []byte{})
    hash := sha256.Sum256(data)
    return hash[:]
}