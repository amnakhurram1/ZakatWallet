package db

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "io"
    "time"
   "wallet_backend_go/internal/models" 
    "wallet_backend_go/internal/blockchain"
)


const (
	tableUsers          = "users"
	tableWalletProfiles = "wallet_profiles"
	tableZakat          = "zakat_records"
	tableSystemLogs     = "system_logs"
)
// SupabaseClient is a minimal client that only knows how to
// talk to Supabase REST using the URL and API key.
type SupabaseClient struct {
    URL string
    Key string
}

// NewSupabaseClient reads SUPABASE_URL and SUPABASE_KEY from the
// environment and returns a SupabaseClient.
func NewSupabaseClient() (*SupabaseClient, error) {
    url := os.Getenv("SUPABASE_URL")
    key := os.Getenv("SUPABASE_KEY")

    if url == "" || key == "" {
        return nil, fmt.Errorf("SUPABASE_URL or SUPABASE_KEY is not set")
    }

    return &SupabaseClient{
        URL: url,
        Key: key,
    }, nil
}

// BlockRecord is the row shape in the "blocks" table.
type BlockRecord struct {
    Hash      string          `json:"hash"`
    Height    int             `json:"height"`
    Timestamp int64           `json:"timestamp"`
    PrevHash  string          `json:"prev_hash"`
    TxCount   int             `json:"tx_count"`
    RawJSON   json.RawMessage `json:"raw_json"`
}

// SaveBlock inserts a block into the Supabase "blocks" table using
// the PostgREST endpoint at /rest/v1/blocks.
func (s *SupabaseClient) SaveBlock(ctx context.Context, height int, block *blockchain.Block) error {
    if s == nil {
        return fmt.Errorf("Supabase client is nil")
    }

    // Serialize full block to JSON for explorer/details
    raw, err := json.Marshal(block)
    if err != nil {
        return fmt.Errorf("marshal block: %w", err)
    }

    rec := BlockRecord{
        Hash:      fmt.Sprintf("%x", block.Hash),
        Height:    height,
        Timestamp: block.Timestamp,
        PrevHash:  fmt.Sprintf("%x", block.PrevHash),
        TxCount:   len(block.Transactions),
        RawJSON:   raw,
    }

    payload, err := json.Marshal([]BlockRecord{rec}) // Supabase expects an array
    if err != nil {
        return fmt.Errorf("marshal payload: %w", err)
    }

    url := fmt.Sprintf("%s/rest/v1/blocks", s.URL)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
    if err != nil {
        return fmt.Errorf("new request: %w", err)
    }

    req.Header.Set("apikey", s.Key)
    req.Header.Set("Authorization", "Bearer "+s.Key)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=minimal")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("supabase insert failed: %s", resp.Status)
    }

    return nil
}


// TransactionRecord is the row shape in the "transactions" table.
type TransactionRecord struct {
    TxID      string          `json:"txid"`
    BlockHash string          `json:"block_hash"`
    Sender    string          `json:"sender"`
    Receiver  string          `json:"receiver"`
    Amount    int             `json:"amount"`
    Timestamp int64           `json:"timestamp"`
    Type      string          `json:"type"` // e.g. "send", "reward", "zakat"
    RawJSON   json.RawMessage `json:"raw_json"`
}



// SaveTransaction inserts a transaction into the Supabase "transactions" table.
func (s *SupabaseClient) SaveTransaction(
    ctx context.Context,
    blockHash string,
    tx *blockchain.Transaction,
    sender string,
    receiver string,
    amount int,
    txType string,
) error {
    if s == nil {
        return fmt.Errorf("Supabase client is nil")
    }

    raw, err := json.Marshal(tx)
    if err != nil {
        return fmt.Errorf("marshal tx: %w", err)
    }

    rec := TransactionRecord{
        TxID:      fmt.Sprintf("%x", tx.ID),
        BlockHash: blockHash,
        Sender:    sender,
        Receiver:  receiver,
        Amount:    amount,
        Timestamp: time.Now().Unix(),
        Type:      txType,
        RawJSON:   raw,
    }

    payload, err := json.Marshal([]TransactionRecord{rec}) // Supabase expects an array
    if err != nil {
        return fmt.Errorf("marshal payload: %w", err)
    }

    url := fmt.Sprintf("%s/rest/v1/transactions", s.URL)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
    if err != nil {
        return fmt.Errorf("new request: %w", err)
    }

    req.Header.Set("apikey", s.Key)
    req.Header.Set("Authorization", "Bearer "+s.Key)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=minimal")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("supabase tx insert failed: %s", resp.Status)
    }

    return nil
}

// CreateUser inserts a new user row.
func (c *SupabaseClient) CreateUser(ctx context.Context, user *models.User) error {
	if c == nil {
		return nil // no-op if Supabase not configured
	}

	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/rest/v1/%s", c.URL, tableUsers),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	// Prefer: return inserted object
	req.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("supabase CreateUser error: %s", resp.Status)
	}
	return nil
}

// CreateWalletProfile inserts wallet info for a user.
func (c *SupabaseClient) CreateWalletProfile(ctx context.Context, wp *models.WalletProfile) error {
	if c == nil {
		return nil
	}

	payload, err := json.Marshal(wp)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/rest/v1/%s", c.URL, tableWalletProfiles),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("supabase CreateWalletProfile error: %s", resp.Status)
	}
	return nil
}

// LogSystemEvent writes a simple log row.
func (c *SupabaseClient) LogSystemEvent(ctx context.Context, level, typ, message, ip string) {
	if c == nil {
		return
	}

	log := models.SystemLog{
		Level:     level,
		Type:      typ,
		Message:   message,
		IP:        ip,
		Timestamp: time.Now().UTC(),
	}

	payload, err := json.Marshal(log)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/rest/v1/%s", c.URL, tableSystemLogs),
		bytes.NewReader(payload),
	)
	if err != nil {
		return
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	_, _ = http.DefaultClient.Do(req) // fire-and-forget
}

// SaveZakatRecord inserts zakat deduction info.
func (c *SupabaseClient) SaveZakatRecord(ctx context.Context, zr *models.ZakatRecord) error {
	if c == nil {
		return nil
	}

	payload, err := json.Marshal(zr)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/rest/v1/%s", c.URL, tableZakat),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.Key)
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("supabase SaveZakatRecord error: %s", resp.Status)
	}
	return nil
}


// ListZakatByWallet returns all zakat_records for a given wallet.
func (c *SupabaseClient) ListZakatByWallet(ctx context.Context, address string) ([]models.ZakatRecord, error) {
    if c == nil {
        return nil, fmt.Errorf("supabase client is nil")
    }

    url := fmt.Sprintf("%s/rest/v1/%s?select=*&wallet_address=eq.%s", c.URL, tableZakat, address)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("apikey", c.Key)
    req.Header.Set("Authorization", "Bearer "+c.Key)
    req.Header.Set("Accept", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("supabase ListZakatByWallet error: %s - %s", resp.Status, string(body))
    }

    var records []models.ZakatRecord
    if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
        return nil, err
    }

    return records, nil
}
// ListSystemLogs returns the most recent system log entries, ordered by timestamp desc.
func (c *SupabaseClient) ListSystemLogs(ctx context.Context, limit int) ([]models.SystemLog, error) {
    if c == nil {
        return nil, fmt.Errorf("supabase client is nil")
    }
    if limit <= 0 {
        limit = 100
    }

    url := fmt.Sprintf("%s/rest/v1/%s?select=*&order=timestamp.desc&limit=%d", c.URL, tableSystemLogs, limit)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("apikey", c.Key)
    req.Header.Set("Authorization", "Bearer "+c.Key)
    req.Header.Set("Accept", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("supabase ListSystemLogs error: %s - %s", resp.Status, string(body))
    }

    var logs []models.SystemLog
    if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
        return nil, err
    }

    return logs, nil
}


// ListTransactionsByWallet returns all transactions where the given wallet
// address is either the sender or the receiver.
func (c *SupabaseClient) ListTransactionsByWallet(ctx context.Context, address string) ([]TransactionRecord, error) {
    if c == nil {
        return nil, fmt.Errorf("supabase client is nil")
    }

    // PostgREST OR filter: sender == address OR receiver == address
    url := fmt.Sprintf("%s/rest/v1/transactions?select=*&or=(sender.eq.%s,receiver.eq.%s)", c.URL, address, address)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("apikey", c.Key)
    req.Header.Set("Authorization", "Bearer "+c.Key)
    req.Header.Set("Accept", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("supabase ListTransactionsByWallet error: %s - %s", resp.Status, string(body))
    }

    var records []TransactionRecord
    if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
        return nil, err
    }

    return records, nil
}



// ListWalletProfiles fetches all wallet_profiles from Supabase.
func (c *SupabaseClient) ListWalletProfiles(ctx context.Context) ([]models.WalletProfile, error) {
    if c == nil {
        return nil, fmt.Errorf("supabase client is nil")
    }

    // Basic: select all columns from wallet_profiles
    url := fmt.Sprintf("%s/rest/v1/%s?select=*", c.URL, tableWalletProfiles)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("apikey", c.Key)
    req.Header.Set("Authorization", "Bearer "+c.Key)
    req.Header.Set("Accept", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("supabase ListWalletProfiles error: %s - %s", resp.Status, string(body))
    }

    var profiles []models.WalletProfile
    if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
        return nil, err
    }

    return profiles, nil
}
