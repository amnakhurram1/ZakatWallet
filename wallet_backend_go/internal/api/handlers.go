package api

// handlers.go defines HTTP handler functions for the REST API. The
// handlers construct JSON responses and invoke the underlying
// blockchain primitives. Errors are returned with a HTTP 400 status.

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
     "sync"
     "crypto/rand"
     "math/big"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"wallet_backend_go/internal/blockchain"
	"wallet_backend_go/internal/db"
	"wallet_backend_go/internal/models"
)

// Server encapsulates the blockchain and its UTXO set. It exposes
// methods that implement the REST API for wallet creation,
// querying balances and sending transactions.
type otpEntry struct {
    Code    string
    Expires time.Time
}

type Server struct {
    BC   *blockchain.Blockchain
    UTXO *blockchain.UTXOSet
    DB   *db.SupabaseClient

    otpMu sync.Mutex
    otps  map[string]otpEntry // key = email
}

type walletReportResponse struct {
    WalletAddress string                `json:"wallet_address"`
    Balance       int                   `json:"balance"`
    TotalSent     int                   `json:"total_sent"`
    TotalReceived int                   `json:"total_received"`
    TotalZakat    int                   `json:"total_zakat"`
    Transactions  []db.TransactionRecord `json:"transactions"`
    ZakatRecords  []models.ZakatRecord  `json:"zakat_records"`
}

type systemLogsResponse struct {
    Logs []models.SystemLog `json:"logs"`
}


// NewServer constructs a Server with the provided blockchain. It
// initializes the UTXO set wrapper around the blockchain and tries
// to create a Supabase client. If Supabase env vars are missing,
// DB will be nil and the API will still work in-memory.
func NewServer(bc *blockchain.Blockchain) *Server {
	var supa *db.SupabaseClient

	client, err := db.NewSupabaseClient()
	if err != nil {
		log.Printf("warning: could not initialize Supabase client: %v", err)
		supa = nil
	} else {
		supa = client
		log.Println("Supabase client initialized")
	}

	return &Server{
		BC:   bc,
		UTXO: &blockchain.UTXOSet{BC: bc},
		DB:   supa,
        otps: make(map[string]otpEntry),
	}
}

// Health responds with a simple JSON object indicating service
// availability.
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// CreateWallet generates a new wallet (private/public key pair) and
// returns its address and private key as hex strings. In a real
// application you would not return the raw private key; instead you
// would prompt the user to securely store it client side.
func (s *Server) CreateWallet(w http.ResponseWriter, r *http.Request) {
	wallet := blockchain.NewWallet()
	resp := map[string]string{
		"address":     wallet.GetAddress(),
		"private_key": hex.EncodeToString(wallet.PrivateKey.D.Bytes()),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// helper: compute balance + pubKeyHash for an address
func (s *Server) balanceForAddress(address string) (int, []byte, error) {
	if !blockchain.ValidateAddress(address) {
		return 0, nil, fmt.Errorf("invalid address")
	}

	pubKeyHash, err := hex.DecodeString(address)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid address")
	}

	UTXOs := s.BC.FindUTXO(pubKeyHash)
	balance := 0
	for _, outs := range UTXOs {
		for _, out := range outs {
			if string(out.PubKeyHash) == string(pubKeyHash) {
				balance += out.Value
			}
		}
	}

	return balance, pubKeyHash, nil
}


func generateOTP(length int) (string, error) {
    result := ""
    for i := 0; i < length; i++ {
        n, err := rand.Int(rand.Reader, big.NewInt(10))
        if err != nil {
            return "", err
        }
        result += fmt.Sprintf("%d", n.Int64())
    }
    return result, nil
}

func (s *Server) WalletReport(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    address := vars["address"]

    if address == "" {
        http.Error(w, "address is required", http.StatusBadRequest)
        return
    }

    if s.DB == nil {
        http.Error(w, "database not configured", http.StatusInternalServerError)
        return
    }

     balance, _, err := s.balanceForAddress(address)
    if err != nil {
        http.Error(w, "invalid address", http.StatusBadRequest)
        return
    }

    // 2) Transactions involving this wallet
    txs, err := s.DB.ListTransactionsByWallet(ctx, address)
    if err != nil {
        http.Error(w, "failed to list transactions", http.StatusInternalServerError)
        s.DB.LogSystemEvent(ctx, "error", "wallet_report_list_txs_failed", err.Error(), r.RemoteAddr)
        return
    }

    // 3) Compute total sent/received from the tx records
    totalSent := 0
    totalReceived := 0
    for _, t := range txs {
        if t.Sender == address {
            totalSent += t.Amount
        }
        if t.Receiver == address {
            totalReceived += t.Amount
        }
    }

    // 4) Zakat records for this wallet
    zakatRecords, err := s.DB.ListZakatByWallet(ctx, address)
    if err != nil {
        http.Error(w, "failed to list zakat records", http.StatusInternalServerError)
        s.DB.LogSystemEvent(ctx, "error", "wallet_report_list_zakat_failed", err.Error(), r.RemoteAddr)
        return
    }

    totalZakat := 0
    for _, zr := range zakatRecords {
        totalZakat += zr.Amount
    }

    resp := walletReportResponse{
        WalletAddress: address,
        Balance:       balance,
        TotalSent:     totalSent,
        TotalReceived: totalReceived,
        TotalZakat:    totalZakat,
        Transactions:  txs,
        ZakatRecords:  zakatRecords,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (s *Server) SystemLogs(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    if s.DB == nil {
        http.Error(w, "database not configured", http.StatusInternalServerError)
        return
    }

    // Optional: limit query parameter
    limit := 100
    if l := r.URL.Query().Get("limit"); l != "" {
        var parsed int
        if _, err := fmt.Sscanf(l, "%d", &parsed); err == nil && parsed > 0 && parsed <= 1000 {
            limit = parsed
        }
    }

    logs, err := s.DB.ListSystemLogs(ctx, limit)
    if err != nil {
        http.Error(w, "failed to list system logs", http.StatusInternalServerError)
        s.DB.LogSystemEvent(ctx, "error", "system_logs_list_failed", err.Error(), r.RemoteAddr)
        return
    }

    resp := systemLogsResponse{
        Logs: logs,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// GetBalance returns the wallet's balance by summing all UTXOs
// belonging to the provided address. The address is extracted from
// the URL path. If the address is invalid or no balance is found,
// zero is returned.
func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	balance, _, err := s.balanceForAddress(address)
	if err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

type registerRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	CNIC     string `json:"cnic"`
}

type registerResponse struct {
	UserID        string `json:"user_id"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	CNIC          string `json:"cnic"`
	WalletAddress string `json:"wallet_address"`
	// For demo / assignment only — in real life you NEVER return this
	PrivateKey string `json:"private_key"`
}

type fundWalletRequest struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

type fundWalletResponse struct {
	Address   string `json:"address"`
	Amount    int    `json:"amount"`
	BlockHash string `json:"block_hash"`
}


type requestOTPRequest struct {
    Email string `json:"email"`
}

type requestOTPResponse struct {
    Email string `json:"email"`
    OTP   string `json:"otp"` // in real life you would NOT return this
}

type verifyOTPRequest struct {
    Email string `json:"email"`
    OTP   string `json:"otp"`
}

type verifyOTPResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

// txRequest defines the payload expected in a send transaction request.
// From and To are addresses as hex strings; Amount is an integer
// number of units to send; PrivKey is the sender's private key as a
// hex encoded big integer (the D value of the ECDSA private key).
type txRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  int    `json:"amount"`
	PrivKey string `json:"privKey"`
}


func (s *Server) RequestOTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req requestOTPRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    if req.Email == "" {
        http.Error(w, "email is required", http.StatusBadRequest)
        return
    }

    code, err := generateOTP(6)
    if err != nil {
        http.Error(w, "failed to generate otp", http.StatusInternalServerError)
        return
    }

    s.otpMu.Lock()
    s.otps[req.Email] = otpEntry{
        Code:    code,
        Expires: time.Now().Add(5 * time.Minute),
    }
    s.otpMu.Unlock()

    if s.DB != nil {
        s.DB.LogSystemEvent(ctx, "info", "otp_generated",
            fmt.Sprintf("otp generated for email=%s", req.Email),
            r.RemoteAddr,
        )
    }

    // In a real app, you would send this via email.
    // For the project/demo, returning it in JSON is enough to show OTP flow.
    resp := requestOTPResponse{
        Email: req.Email,
        OTP:   code,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}


func (s *Server) VerifyOTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req verifyOTPRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    if req.Email == "" || req.OTP == "" {
        http.Error(w, "email and otp are required", http.StatusBadRequest)
        return
    }

    s.otpMu.Lock()
    entry, ok := s.otps[req.Email]
    s.otpMu.Unlock()

    if !ok {
        if s.DB != nil {
            s.DB.LogSystemEvent(ctx, "warn", "otp_not_found",
                fmt.Sprintf("no otp for email=%s", req.Email),
                r.RemoteAddr,
            )
        }
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(verifyOTPResponse{
            Success: false,
            Message: "invalid or expired otp",
        })
        return
    }

    if time.Now().After(entry.Expires) || entry.Code != req.OTP {
        if s.DB != nil {
            s.DB.LogSystemEvent(ctx, "warn", "otp_invalid",
                fmt.Sprintf("invalid otp for email=%s", req.Email),
                r.RemoteAddr,
            )
        }
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(verifyOTPResponse{
            Success: false,
            Message: "invalid or expired otp",
        })
        return
    }

    // OTP valid – consider the user "authenticated" for this demo.
    if s.DB != nil {
        s.DB.LogSystemEvent(ctx, "info", "otp_verified",
            fmt.Sprintf("otp verified for email=%s", req.Email),
            r.RemoteAddr,
        )
    }

    // Optionally: delete OTP so it can't be reused
    s.otpMu.Lock()
    delete(s.otps, req.Email)
    s.otpMu.Unlock()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(verifyOTPResponse{
        Success: true,
        Message: "otp verified",
    })
}

// SendTransaction constructs, signs and broadcasts a new transaction.
// It expects a JSON body containing from, to, amount and privKey.
// The transaction is mined into a new block immediately for
// demonstration purposes. Errors in decoding or signing are
// reported with HTTP 400.
func (s *Server) SendTransaction(w http.ResponseWriter, r *http.Request) {
	var req txRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	if !blockchain.ValidateAddress(req.From) || !blockchain.ValidateAddress(req.To) {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}
	// decode private key big integer
	dBytes, err := hex.DecodeString(req.PrivKey)
	if err != nil {
		http.Error(w, "invalid private key", http.StatusBadRequest)
		return
	}
	// reconstruct ECDSA private key
	curve := blockchain.GetDefaultCurve()
	priv := blockchain.BigIntToPrivateKey(dBytes, curve)
	// find spendable outputs
	fromPubKeyHash, _ := hex.DecodeString(req.From)
	amount, spendable := s.UTXO.FindSpendableOutputs(fromPubKeyHash, req.Amount)
	if amount < req.Amount {
		http.Error(w, "insufficient funds", http.StatusBadRequest)
		return
	}
	// build transaction
	tx, err := blockchain.NewUTXOTransaction(priv, req.To, req.Amount, s.BC, spendable, fromPubKeyHash, amount)
	if err != nil {
		http.Error(w, "failed to create transaction", http.StatusBadRequest)
		return
	}
	// verify transaction before adding
	if !s.BC.VerifyTransaction(tx) {
		http.Error(w, "invalid transaction", http.StatusBadRequest)
		return
	}

	// mine new block
	newBlock := s.BC.AddBlock([]*blockchain.Transaction{tx})

	// persist block + transaction to Supabase (if DB is configured)
	height := len(s.BC.Blocks) - 1
	if s.DB != nil {
		blockHash := fmt.Sprintf("%x", newBlock.Hash)
		fromAddress := req.From
		toAddress := req.To
		sentAmount := req.Amount

		go func(b *blockchain.Block, h int, bh, from, to string, amt int, tx *blockchain.Transaction) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// save block
			if err := s.DB.SaveBlock(ctx, h, b); err != nil {
				log.Printf("failed to save block to Supabase: %v", err)
			}

			// save transaction
			if err := s.DB.SaveTransaction(ctx, bh, tx, from, to, amt, "send"); err != nil {
				log.Printf("failed to save transaction to Supabase: %v", err)
			}
		}(newBlock, height, blockHash, fromAddress, toAddress, sentAmount, tx)
	}

	// update UTXO set
	_ = s.UTXO.Reindex()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "transaction mined"})
}

// ListBlocks returns a summary of all blocks in the chain.
func (s *Server) ListBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	summaries := s.BC.ListBlocks()
	_ = json.NewEncoder(w).Encode(summaries)
}

// GetBlock returns the full block at the given index.
func (s *Server) GetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idxStr := vars["index"]

	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		http.Error(w, "invalid block index", http.StatusBadRequest)
		return
	}

	block, ok := s.BC.GetBlockByIndex(idx)
	if !ok {
		http.Error(w, "block not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(block)
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.FullName == "" || req.Email == "" || req.CNIC == "" {
		http.Error(w, "full_name, email and cnic are required", http.StatusBadRequest)
		return
	}

	// 1) Create blockchain wallet (using your existing wallet logic)
	wallet := blockchain.NewWallet()
	address := wallet.GetAddress()

	// Convert keys to hex strings
	privKeyHex := blockchain.PrivateKeyToHex(&wallet.PrivateKey)
	pubKeyHex := hex.EncodeToString(wallet.PublicKey)

	// "Encrypt" private key (for assignment we can just base64 it)
	encryptedPriv := base64.StdEncoding.EncodeToString([]byte(privKeyHex))

	// 2) Create user record
	user := &models.User{
		ID:        uuid.NewString(),
		FullName:  req.FullName,
		Email:     req.Email,
		CNIC:      req.CNIC,
		CreatedAt: time.Now().UTC(),
	}

	if s.DB != nil {
		if err := s.DB.CreateUser(ctx, user); err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			if s.DB != nil {
				s.DB.LogSystemEvent(ctx, "error", "user_create_failed", err.Error(), r.RemoteAddr)
			}
			return
		}

		// 3) Create wallet profile linked to the user
		wp := &models.WalletProfile{
			ID:                  uuid.NewString(),
			UserID:              user.ID,
			WalletAddress:       address,
			PublicKeyHex:        pubKeyHex,
			EncryptedPrivateKey: encryptedPriv,
			CreatedAt:           time.Now().UTC(),
		}

		if err := s.DB.CreateWalletProfile(ctx, wp); err != nil {
			http.Error(w, "failed to create wallet profile", http.StatusInternalServerError)
			if s.DB != nil {
				s.DB.LogSystemEvent(ctx, "error", "wallet_profile_create_failed", err.Error(), r.RemoteAddr)
			}
			return
		}

		s.DB.LogSystemEvent(ctx, "info", "user_registered",
			fmt.Sprintf("user %s registered with wallet %s", user.Email, address),
			r.RemoteAddr,
		)
	}

	// 4) Send response (including private key so user can use wallet)
	resp := registerResponse{
		UserID:        user.ID,
		FullName:      user.FullName,
		Email:         user.Email,
		CNIC:          user.CNIC,
		WalletAddress: address,
		PrivateKey:    privKeyHex, // show raw hex for now so they can sign tx
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Zakat run response
type zakatRunResponse struct {
	TotalWallets int      `json:"total_wallets"`
	Processed    int      `json:"processed"`
	TotalZakat   int      `json:"total_zakat"`
	BlockHashes  []string `json:"block_hashes"`
}

// RunZakat calculates 2.5% zakat for each wallet and sends it to the Zakat pool wallet.
func (s *Server) RunZakat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.DB == nil {
		http.Error(w, "database not configured", http.StatusInternalServerError)
		return
	}

	zakatAddress := os.Getenv("ZAKAT_WALLET_ADDRESS")
	if zakatAddress == "" {
		http.Error(w, "ZAKAT_WALLET_ADDRESS not set", http.StatusInternalServerError)
		return
	}

	// 1) Fetch all wallet profiles from Supabase
	profiles, err := s.DB.ListWalletProfiles(ctx)
	if err != nil {
		http.Error(w, "failed to list wallet profiles", http.StatusInternalServerError)
		s.DB.LogSystemEvent(ctx, "error", "zakat_list_wallets_failed", err.Error(), r.RemoteAddr)
		return
	}

	processed := 0
	totalZakat := 0
	var blockHashes []string

	for _, wp := range profiles {
		addr := wp.WalletAddress

		// compute balance
		balance, pubKeyHash, balErr := s.balanceForAddress(addr)
		if balErr != nil || balance <= 0 {
			if balErr != nil {
				s.DB.LogSystemEvent(ctx, "error", "zakat_balance_failed", balErr.Error(), r.RemoteAddr)
			}
			continue
		}

		// zakat = 2.5% => balance * 25 / 1000
		zakatAmount := (balance * 25) / 1000
		if zakatAmount <= 0 {
			continue
		}

		// Decode "encrypted" private key (base64 of hex string)
		decoded, decErr := base64.StdEncoding.DecodeString(wp.EncryptedPrivateKey)
		if decErr != nil {
			s.DB.LogSystemEvent(ctx, "error", "zakat_privkey_decode_failed", decErr.Error(), r.RemoteAddr)
			continue
		}

		privHex := string(decoded)
		privKey, pkErr := blockchain.PrivateKeyFromHex(privHex)
		if pkErr != nil {
			s.DB.LogSystemEvent(ctx, "error", "zakat_privkey_reconstruct_failed", pkErr.Error(), r.RemoteAddr)
			continue
		}

		// Find spendable outputs for zakat amount
		amount, spendable := s.UTXO.FindSpendableOutputs(pubKeyHash, zakatAmount)
		if amount < zakatAmount {
			// not enough balance in UTXOs (should not normally happen if balance check is correct)
			continue
		}

		// Create zakat transaction
		tx, txErr := blockchain.NewUTXOTransaction(*privKey, zakatAddress, zakatAmount, s.BC, spendable, pubKeyHash, amount)
		if txErr != nil {
			s.DB.LogSystemEvent(ctx, "error", "zakat_tx_create_failed", txErr.Error(), r.RemoteAddr)
			continue
		}

		// Verify transaction
		if !s.BC.VerifyTransaction(tx) {
			s.DB.LogSystemEvent(ctx, "error", "zakat_tx_verify_failed", "verification failed", r.RemoteAddr)
			continue
		}

		// Mine block with this zakat transaction
		newBlock := s.BC.AddBlock([]*blockchain.Transaction{tx})
		blockHashHex := fmt.Sprintf("%x", newBlock.Hash)
		blockHashes = append(blockHashes, blockHashHex)
		processed++
		totalZakat += zakatAmount

		// Update UTXO set (rebuild)
		_ = s.UTXO.Reindex()

		// Save block & transaction as zakat_deduction
		height := len(s.BC.Blocks) - 1
		if saveBlkErr := s.DB.SaveBlock(ctx, height, newBlock); saveBlkErr != nil {
			s.DB.LogSystemEvent(ctx, "error", "zakat_block_save_failed", saveBlkErr.Error(), r.RemoteAddr)
		}

		if saveTxErr := s.DB.SaveTransaction(ctx, blockHashHex, tx, addr, zakatAddress, zakatAmount, "zakat_deduction"); saveTxErr != nil {
			s.DB.LogSystemEvent(ctx, "error", "zakat_tx_save_failed", saveTxErr.Error(), r.RemoteAddr)
		}

		// Save zakat record
		zr := &models.ZakatRecord{
			ID:            uuid.NewString(),
			UserID:        wp.UserID,
			WalletAddress: addr,
			Amount:        zakatAmount,
			BlockHash:     blockHashHex,
			CreatedAt:     time.Now().UTC(),
		}
		if zrErr := s.DB.SaveZakatRecord(ctx, zr); zrErr != nil {
			s.DB.LogSystemEvent(ctx, "error", "zakat_record_save_failed", zrErr.Error(), r.RemoteAddr)
		}
	}

	s.DB.LogSystemEvent(ctx, "info", "zakat_run",
		fmt.Sprintf("zakat run processed=%d total_zakat=%d", processed, totalZakat),
		r.RemoteAddr,
	)

	resp := zakatRunResponse{
		TotalWallets: len(profiles),
		Processed:    processed,
		TotalZakat:   totalZakat,
		BlockHashes:  blockHashes,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetWalletTransactions returns all transactions that involve the
// given wallet address as a recipient.
func (s *Server) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	if !blockchain.ValidateAddress(address) {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	txs, err := s.BC.GetTransactionsForAddress(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(txs)
}

// FundWallet: admin faucet to fund a wallet via coinbase transaction.
func (s *Server) FundWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req fundWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Address == "" || req.Amount <= 0 {
		http.Error(w, "address and positive amount are required", http.StatusBadRequest)
		return
	}

	if !blockchain.ValidateAddress(req.Address) {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	// 1) Create coinbase transaction paying to this address
	cbTx := blockchain.NewCoinbaseTx(req.Address, "admin_faucet_reward")

	// 2) Mine block with this coinbase tx
	newBlock := s.BC.AddBlock([]*blockchain.Transaction{cbTx})

	// 3) Rebuild UTXO set
	_ = s.UTXO.Reindex()

	blockHashHex := fmt.Sprintf("%x", newBlock.Hash)

	if s.DB != nil {
		// save block
		if err := s.DB.SaveBlock(ctx, len(s.BC.Blocks)-1, newBlock); err != nil {
			s.DB.LogSystemEvent(ctx, "error", "faucet_save_block_failed", err.Error(), r.RemoteAddr)
		}
		// save tx as reward
		if len(newBlock.Transactions) > 0 {
			if err := s.DB.SaveTransaction(ctx,
				blockHashHex,
				newBlock.Transactions[0],
				"SYSTEM",
				req.Address,
				req.Amount,
				"reward",
			); err != nil {
				s.DB.LogSystemEvent(ctx, "error", "faucet_save_tx_failed", err.Error(), r.RemoteAddr)
			}
		}
		s.DB.LogSystemEvent(ctx, "info", "faucet_fund",
			fmt.Sprintf("funded %d to %s", req.Amount, req.Address),
			r.RemoteAddr,
		)
	}

	resp := fundWalletResponse{
		Address:   req.Address,
		Amount:    req.Amount,
		BlockHash: blockHashHex,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Router sets up route definitions using gorilla/mux. This function returns
// an http.Handler that can be passed to http.ListenAndServe. API
// versioning is prefixed on all routes.
func (s *Server) Router() http.Handler {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/register", s.Register).Methods("POST")
	api.HandleFunc("/health", s.Health).Methods("GET")
	api.HandleFunc("/admin/fund", s.FundWallet).Methods("POST")

    api.HandleFunc("/auth/request-otp", s.RequestOTP).Methods("POST")
api.HandleFunc("/auth/verify-otp", s.VerifyOTP).Methods("POST")


	// Zakat endpoint
	api.HandleFunc("/zakat/run", s.RunZakat).Methods("POST")

	// Wallet endpoints
	api.HandleFunc("/wallets", s.CreateWallet).Methods("POST")
	api.HandleFunc("/wallets/{address}/balance", s.GetBalance).Methods("GET")
	api.HandleFunc("/wallets/{address}/transactions", s.GetWalletTransactions).Methods("GET")

	// Transaction endpoint
	api.HandleFunc("/transactions", s.SendTransaction).Methods("POST")

	// Block explorer endpoints
	api.HandleFunc("/blocks", s.ListBlocks).Methods("GET")
	api.HandleFunc("/blocks/{index}", s.GetBlock).Methods("GET")
	api.HandleFunc("/reports/wallet/{address}", s.WalletReport).Methods("GET")
api.HandleFunc("/logs/system", s.SystemLogs).Methods("GET")


	return r
}
