package blockchain

// wallet.go provides a simple wallet abstraction using ECDSA keys. A
// wallet contains a private/public key pair and can derive an
// address by hashing the public key. Addresses are returned as hex
// strings. In a production system you'd want proper base58
// encoding and checksum validation.

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "fmt"
    "math/big"
    "encoding/hex"
)

// Wallet holds an ECDSA private key and its corresponding public key.
type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey  []byte
}

// NewWallet generates a new ECDSA private/public key pair. It uses
// the P‑256 curve. Any error during key generation will panic,
// although random failures are extremely unlikely.
func NewWallet() *Wallet {
    privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        panic(err)
    }
    pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
    return &Wallet{PrivateKey: *privKey, PublicKey: pubKey}
}

// GetAddress derives a simple address by hashing the public key with
// SHA‑256 and returning the result as a hexadecimal string. Real
// blockchain addresses typically use base58check or bech32 encoding
// with version prefixes and checksums.
func (w *Wallet) GetAddress() string {
    pubHash := sha256.Sum256(w.PublicKey)
    return fmt.Sprintf("%x", pubHash[:])
}

// ValidateAddress performs a basic length check on the address. In
// practice you'd also verify the checksum and prefix.
func ValidateAddress(address string) bool {
    return len(address) > 0
}


// PrivateKeyToHex converts an ECDSA private key to hex string (using D).
func PrivateKeyToHex(priv *ecdsa.PrivateKey) string {
    return fmt.Sprintf("%x", priv.D.Bytes())
}

// PrivateKeyFromHex reconstructs an ECDSA private key from its hex-encoded D value.
func PrivateKeyFromHex(hexKey string) (*ecdsa.PrivateKey, error) {
    dBytes, err := hex.DecodeString(hexKey)
    if err != nil {
        return nil, fmt.Errorf("decode hex private key: %w", err)
    }

    curve := elliptic.P256()
    priv := new(ecdsa.PrivateKey)
    priv.PublicKey.Curve = curve
    priv.D = new(big.Int).SetBytes(dBytes)
    priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(dBytes)

    return priv, nil
}
