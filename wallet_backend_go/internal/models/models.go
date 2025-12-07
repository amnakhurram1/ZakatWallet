package models

import "time"

// User represents an application user (NOT blockchain only).
// This will be stored in a "users" table in Supabase.
type User struct {
	ID        string    `json:"id"`         // uuid in Supabase
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	CNIC      string    `json:"cnic"`       // National ID
	CreatedAt time.Time `json:"created_at"`
}

// WalletProfile links a user to a blockchain wallet.
type WalletProfile struct {
	ID                  string    `json:"id"`                     // uuid
	UserID              string    `json:"user_id"`                // foreign key -> users.id
	WalletAddress       string    `json:"wallet_address"`         // hash of pub key (your existing address)
	PublicKeyHex        string    `json:"public_key_hex"`         // hex-encoded
	EncryptedPrivateKey string    `json:"encrypted_private_key"`  // we'll just store raw for now, can "pretend" it's encrypted
	CreatedAt           time.Time `json:"created_at"`
}

// ZakatRecord stores each zakat deduction operation.
type ZakatRecord struct {
	ID            string    `json:"id"`             // uuid
	UserID        string    `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	Amount        int       `json:"amount"`         // integer amount of "coins"
	BlockHash     string    `json:"block_hash"`
	CreatedAt     time.Time `json:"created_at"`
}

// SystemLog stores system-level log events.
type SystemLog struct {
	ID        string    `json:"id"`        // uuid
	Level     string    `json:"level"`     // info, warn, error
	Type      string    `json:"type"`      // login_attempt, otp_failed, invalid_wallet, rejected_tx, mining_event, zakat_run, etc.
	Message   string    `json:"message"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}
