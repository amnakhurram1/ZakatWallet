package blockchain

// utxo.go defines a simple UTXO set abstraction. The UTXO set is
// responsible for scanning the blockchain and collecting unspent
// outputs, finding spendable outputs for a given public key hash, and
// updating the set when new blocks are mined. In a production
// implementation this would be backed by a database, but here we
// maintain it in memory and leave persistence to the caller.

import (
    "bytes"
    "fmt"
)

// UTXOSet wraps a blockchain and maintains a cache of unspent
// transaction outputs. For simplicity, the set is a map keyed by
// transaction ID hex strings with values being slices of output
// indexes. Consumers of the set should persist it alongside the
// blockchain in a database or external store.
type UTXOSet struct {
    BC *Blockchain
}

// Reindex rebuilds the entire UTXO set by scanning all blocks. It
// discards any existing cache and reconstructs it from scratch. This
// method should be called when the blockchain is first opened from
// persistent storage. The returned map is keyed by transaction ID
// encoded in hexadecimal, with values being slices of TxOutput.
func (u *UTXOSet) Reindex() map[string][]TxOutput {
    UTXO := make(map[string][]TxOutput)
    if u.BC == nil {
        return UTXO
    }
    unspent := u.BC.FindUTXO(nil)
    for txID, outs := range unspent {
        UTXO[txID] = outs
    }
    return UTXO
}

// FindSpendableOutputs locates enough outputs to cover the given amount.
// It returns the accumulated value and a map of transaction IDs to
// output indexes. pubKeyHash identifies the outputs belonging to the
// requester. This method iterates over the set and stops once the
// accumulated value meets or exceeds the amount.
func (u *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
    accumulated := 0
    unspentOuts := make(map[string][]int)

    // scan the entire blockchain for UTXOs owned by pubKeyHash
    UTXO := u.BC.FindUTXO(pubKeyHash)
    for txID, outs := range UTXO {
        for outIdx, out := range outs {
            if bytes.Equal(out.PubKeyHash, pubKeyHash) && accumulated < amount {
                accumulated += out.Value
                unspentOuts[txID] = append(unspentOuts[txID], outIdx)
                if accumulated >= amount {
                    return accumulated, unspentOuts
                }
            }
        }
    }
    return accumulated, unspentOuts
}

// FindUTXO returns all unspent outputs for the provided public key hash.
// It is a thin wrapper over Blockchain.FindUTXO, which scans the
// blockchain and returns a map of transaction IDs to unspent outputs.
func (u *UTXOSet) FindUTXO(pubKeyHash []byte) map[string][]TxOutput {
    return u.BC.FindUTXO(pubKeyHash)
}

// Update processes a new block and removes spent outputs from the
// UTXO set while adding new outputs. In a persistent implementation
// this would modify the on‑disk database. Here we simply adjust the
// provided in‑memory UTXO map. Each input spends an output from a
// previous transaction; that output is removed from the set. Then
// every output in the new block's transactions is added to the set.
func (u *UTXOSet) Update(block *Block, utxo map[string][]TxOutput) {
    for _, tx := range block.Transactions {
        if !tx.IsCoinbase() {
            for _, vin := range tx.Vin {
                // remove spent output
                outs := utxo[fmt.Sprintf("%x", vin.Txid)]
                var updatedOuts []TxOutput
                for outIdx, out := range outs {
                    spent := false
                    for _, inOutIdx := range []int{vin.Vout} {
                        if outIdx == inOutIdx {
                            spent = true
                            break
                        }
                    }
                    if !spent {
                        updatedOuts = append(updatedOuts, out)
                    }
                }
                if len(updatedOuts) == 0 {
                    delete(utxo, fmt.Sprintf("%x", vin.Txid))
                } else {
                    utxo[fmt.Sprintf("%x", vin.Txid)] = updatedOuts
                }
            }
        }
        // add new outputs
        newOutputs := make([]TxOutput, len(tx.Vout))
        copy(newOutputs, tx.Vout)
        utxo[fmt.Sprintf("%x", tx.ID)] = newOutputs
    }
}