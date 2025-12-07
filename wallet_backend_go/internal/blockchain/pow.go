package blockchain

// pow.go implements a simple proof‑of‑work for blocks. The
// difficulty is defined by targetBits. Miners iterate nonce values
// until the resulting SHA‑256 hash of the block header is less than
// the target. This process provides computational work backing
// block issuance.

import (
    "bytes"
    "crypto/sha256"
    "encoding/binary"
    "math/big"
)

const targetBits = 20 // adjust difficulty here; lower numbers make mining easier

// ProofOfWork ties a block to its difficulty target. The target is a
// big integer computed from targetBits.
type ProofOfWork struct {
    block  *Block
    target *big.Int
}

// NewProofOfWork initializes a proof‑of‑work for the given block.
func NewProofOfWork(b *Block) *ProofOfWork {
    target := big.NewInt(1)
    target.Lsh(target, 256-targetBits)
    pow := &ProofOfWork{block: b, target: target}
    return pow
}

// prepareData constructs the byte slice to be hashed from the block
// fields and the given nonce. The ordering of the fields is
// important; changing it will change the PoW algorithm.
func (pow *ProofOfWork) prepareData(nonce int) []byte {
    return bytes.Join(
        [][]byte{
            pow.block.PrevHash,
            pow.block.HashTransactions(),
            IntToHex(pow.block.Timestamp),
            IntToHex(int64(targetBits)),
            IntToHex(int64(nonce)),
        },
        []byte{},
    )
}

// Run performs the proof‑of‑work search. It repeatedly hashes the
// prepared data with incrementing nonce values until a hash less
// than the target is found. It returns the discovered nonce and the
// corresponding hash.
func (pow *ProofOfWork) Run() (int, []byte) {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0

    for {
        data := pow.prepareData(nonce)
        hash = sha256.Sum256(data)
        hashInt.SetBytes(hash[:])
        if hashInt.Cmp(pow.target) == -1 {
            break
        } else {
            nonce++
        }
    }
    return nonce, hash[:]
}

// Validate executes a single hash with the stored nonce and checks
// whether it meets the target. This is useful when receiving blocks
// from peers and ensures they did the work.
func (pow *ProofOfWork) Validate() bool {
    var hashInt big.Int
    data := pow.prepareData(pow.block.Nonce)
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])
    return hashInt.Cmp(pow.target) == -1
}

// IntToHex converts an integer to a byte slice in big‑endian order.
func IntToHex(n int64) []byte {
    buf := new(bytes.Buffer)
    _ = binary.Write(buf, binary.BigEndian, n)
    return buf.Bytes()
}