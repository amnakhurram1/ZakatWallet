package blockchain

// transaction.go defines the data structures and helper methods used to represent
// and sign blockchain transactions. Transactions follow a UTXO model where
// inputs reference previous unspent outputs and outputs carry a value and
// a public‑key hash that must be satisfied by a future spender.

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/gob"
    "encoding/hex"
    "fmt"
    "math/big"
)


// TxInput references a previous output and carries the spender's
// signature and public key. The signature proves ownership of the
// corresponding private key and must satisfy the referenced output's
// public key hash.
type TxInput struct {
    Txid      []byte // transaction ID of the referenced output
    Vout      int    // index of the referenced output
    Signature []byte // ECDSA signature proving ownership
    PubKey    []byte // raw public key of the spender
}

// TxOutput represents a payment to a public key hash. Value is
// denominated in arbitrary units (e.g. satoshis). The PubKeyHash
// encodes the address (often a hashed public key) that must be
// provided to spend this output.
type TxOutput struct {
    Value      int
    PubKeyHash []byte
}

// Transaction bundles one or more inputs and outputs. The ID field is
// derived from the transaction's serialized form and uniquely
// identifies the transaction on chain.
type Transaction struct {
    ID   []byte
    Vin  []TxInput
    Vout []TxOutput
}

// SetID computes and sets the transaction's ID. A gob encoder is used
// to serialize the transaction; then a SHA‑256 hash of the resulting
// bytes becomes the ID. Mutating the transaction after calling
// SetID will change the content but not the ID, so call this only
// once when the transaction is constructed.
func (tx *Transaction) SetID() {
    var encoded bytes.Buffer
    var hash [32]byte

    enc := gob.NewEncoder(&encoded)
    if err := enc.Encode(tx); err != nil {
        panic(err)
    }
    hash = sha256.Sum256(encoded.Bytes())
    tx.ID = hash[:]
}

// NewCoinbaseTx creates a coinbase transaction awarding a fixed
// subsidy to the provided address. Coinbase transactions have a single
// input with an empty Txid and Vout of ‑1. The Signature and PubKey
// fields can carry arbitrary data; here we store a human‑readable
// message describing the reward. Coinbase outputs pay to the
// recipient's address (represented here directly as a byte slice).
func NewCoinbaseTx(to, data string) *Transaction {
    if data == "" {
        data = fmt.Sprintf("Reward to %s", to)
    }

    txin := TxInput{
        Txid:      []byte{},
        Vout:      -1,
        Signature: nil,
        PubKey:    []byte(data),
    }

    // IMPORTANT: store the *decoded* address bytes, same as normal txs
    var pubKeyHash []byte
    if to != "" {
        decoded, err := hex.DecodeString(to)
        if err == nil {
            pubKeyHash = decoded
        } else {
            // fallback – should not happen with normal wallets
            pubKeyHash = []byte(to)
        }
    }

    txout := TxOutput{
        Value:      15000,
        PubKeyHash: pubKeyHash,
    }

    tx := Transaction{
        ID:   nil,
        Vin:  []TxInput{txin},
        Vout: []TxOutput{txout},
    }
    tx.SetID()
    return &tx
}


// IsCoinbase returns true if the transaction has the structure of a
// coinbase transaction.
func (tx *Transaction) IsCoinbase() bool {
    return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// TrimmedCopy returns a copy of the transaction with blanked out
// signatures and public keys. This copy is used to calculate
// deterministic hashes for signing inputs. Each input's PubKey field
// will later be filled with the previous output's PubKeyHash before
// hashing.
func (tx *Transaction) TrimmedCopy() Transaction {
    var inputs []TxInput
    var outputs []TxOutput

    for _, vin := range tx.Vin {
        inputs = append(inputs, TxInput{Txid: vin.Txid, Vout: vin.Vout, Signature: nil, PubKey: nil})
    }
    for _, vout := range tx.Vout {
        outputs = append(outputs, TxOutput{Value: vout.Value, PubKeyHash: vout.PubKeyHash})
    }

    txCopy := Transaction{ID: tx.ID, Vin: inputs, Vout: outputs}
    return txCopy
}

// Sign signs each input of the transaction using the provided
// private key. prevTXs maps transaction IDs (as hex strings) to
// previous transactions referenced by this transaction. For each input,
// the corresponding previous output's PubKeyHash is injected into
// the trimmed copy, hashed, and then signed. The resulting signature
// is stored in the original transaction's input.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
    if tx.IsCoinbase() {
        return nil
    }

    txCopy := tx.TrimmedCopy()

    for inIdx, vin := range tx.Vin {
        prevTx, ok := prevTXs[fmt.Sprintf("%x", vin.Txid)]
        if !ok {
            return fmt.Errorf("previous transaction not found")
        }
        // Set the referenced output's pubKeyHash on the copy
        txCopy.Vin[inIdx].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
        // Compute hash for signing
        txCopy.ID = txCopy.Hash()
        // Clear the pubkey so the next input doesn't reuse it
        txCopy.Vin[inIdx].PubKey = nil

        r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
        if err != nil {
            return err
        }
        signature := append(r.Bytes(), s.Bytes()...)
        tx.Vin[inIdx].Signature = signature
        tx.Vin[inIdx].PubKey = append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
    }
    return nil
}

// Verify verifies each input's signature against the corresponding
// previous output's PubKeyHash. A copy of the transaction with
// signatures blanked out is used to compute the hash. If any
// signature fails verification, the transaction is invalid.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
    if tx.IsCoinbase() {
        return true
    }

    txCopy := tx.TrimmedCopy()
    curve := elliptic.P256()

    for inIdx, vin := range tx.Vin {
        prevTx := prevTXs[fmt.Sprintf("%x", vin.Txid)]
        // Inject referenced output's pubKeyHash
        txCopy.Vin[inIdx].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
        // Hash for verification
        txCopy.ID = txCopy.Hash()
        // Restore blank pubKey
        txCopy.Vin[inIdx].PubKey = nil

        // Split signature
        r := big.Int{}
        s := big.Int{}
        sigLen := len(vin.Signature)
        r.SetBytes(vin.Signature[:sigLen/2])
        s.SetBytes(vin.Signature[sigLen/2:])

        // Split public key
        x := big.Int{}
        y := big.Int{}
        keyLen := len(vin.PubKey)
        if keyLen == 0 {
            return false
        }
        x.SetBytes(vin.PubKey[:keyLen/2])
        y.SetBytes(vin.PubKey[keyLen/2:])
        rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

        if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
            return false
        }
    }
    return true
}

// Hash returns the SHA‑256 hash of the transaction without its ID. The
// ID field is blanked before hashing to avoid self‑reference. The
// serialization uses gob encoding. This function is used by Sign
// and Verify to generate deterministic hashes.
func (tx Transaction) Hash() []byte {
    var hash [32]byte
    txCopy := tx
    txCopy.ID = []byte{}
    hash = sha256.Sum256(txCopy.Serialize())
    return hash[:]
}

// Serialize encodes the transaction into bytes using gob. It panics
// if encoding fails, as serialization should never fail for well
// defined structs.
func (tx Transaction) Serialize() []byte {
    var encoded bytes.Buffer
    enc := gob.NewEncoder(&encoded)
    if err := enc.Encode(tx); err != nil {
        panic(err)
    }
    return encoded.Bytes()
}