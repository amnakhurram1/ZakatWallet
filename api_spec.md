# REST API Specification for Wallet Backend

This document describes all REST API endpoints exposed by the Go‑based wallet backend.  The server listens on port `8080` and versions all routes under `/api/v1`.  Clients should prefix every path with this base URL when making requests (e.g. `http://localhost:8080/api/v1/register`).  All request and response bodies are JSON‑encoded.  Errors are returned as plain text or JSON along with the appropriate HTTP status code.

## Environment Variables

Several endpoints rely on environment variables being set when the server starts.  When these are missing the affected endpoints will return a 500 error.

| Variable                | Description                                                                    |
|-------------------------|--------------------------------------------------------------------------------|
| `SUPABASE_URL`          | The Supabase REST API base URL used by the database client.                    |
| `SUPABASE_KEY`          | API key for the Supabase instance.                                            |
| `ZAKAT_WALLET_ADDRESS`  | Address of the central Zakat pool wallet; required for `/zakat/run` endpoint. |

If `SUPABASE_URL` or `SUPABASE_KEY` are not defined, the API will run in memory; calls that require the database will return `500 Internal Server Error` with the message `"database not configured"`.

## Health

### `GET /health`

Returns a simple health check indicating that the service is up.

**Response:**

```json
{
  "status": "ok"
}
```

## User Registration

### `POST /register`

Creates a new user record, generates a blockchain wallet for them and returns user details along with the raw private key.  In a real implementation the private key would **never** be sent back to the client; this is done here for demonstration purposes.

**Request Body:**

```json
{
  "full_name": "string",   // required
  "email": "string",       // required
  "cnic": "string"         // required, National ID
}
```

**Successful Response (`200 OK`):**

```json
{
  "user_id": "string",        // UUID of the created user
  "full_name": "string",
  "email": "string",
  "cnic": "string",
  "wallet_address": "string", // hex‑encoded SHA‑256 of the public key
  "private_key": "string"     // hex‑encoded ECDSA private key (for demo only)
}
```

**Errors:**

| Status | Condition                                                   | Response                         |
|-------:|-------------------------------------------------------------|----------------------------------|
| 400    | Malformed JSON or missing `full_name`, `email` or `cnic`    | Plain text message               |
| 500    | Database insert fails (only when Supabase configured)       | Plain text message               |

## Authentication (OTP)

The OTP flow is used to simulate a login mechanism.  OTPs are generated and stored in memory; they expire after 5 minutes.

### `POST /auth/request-otp`

Generates a one‑time password for the supplied email and returns it directly in the response.  A real application would instead email the OTP to the user.

**Request Body:**

```json
{
  "email": "string"  // required
}
```

**Successful Response (`200 OK`):**

```json
{
  "email": "string",
  "otp": "string"    // 6‑digit numerical code
}
```

**Errors:**

| Status | Condition                    | Response            |
|-------:|------------------------------|---------------------|
| 400    | Invalid JSON or empty email | Plain text message  |
| 500    | Random number generation failed | Plain text message  |

### `POST /auth/verify-otp`

Verifies the one‑time password for the supplied email.  OTPs are removed after successful verification and cannot be reused.

**Request Body:**

```json
{
  "email": "string", // required
  "otp": "string"    // required
}
```

**Successful Response (`200 OK`):**

```json
{
  "success": true,
  "message": "otp verified"
}
```

**Failure Response (`401 Unauthorized`):**

```json
{
  "success": false,
  "message": "invalid or expired otp"
}
```

**Errors:**

| Status | Condition                                      | Response                             |
|-------:|------------------------------------------------|--------------------------------------|
| 400    | Invalid JSON or missing `email`/`otp`          | Plain text message                   |
| 401    | OTP not found, expired or does not match       | JSON body (see above)                |

## Wallet Operations

### `POST /wallets`

Creates a new blockchain wallet (ECDSA key pair) and returns its address and private key.

**Request Body:** *none*

**Successful Response (`200 OK`):**

```json
{
  "address": "string",      // hex‑encoded pubKeyHash
  "private_key": "string"  // hex‑encoded private key (D component)
}
```

### `GET /wallets/{address}/balance`

Returns the current confirmed balance for the specified wallet address.  Balance is calculated by summing all unspent transaction outputs.

**Path Parameters:**

| Name    | Type   | Description                                          |
|---------|--------|------------------------------------------------------|
| address | string | Hex‑encoded hash of the public key (wallet address) |

**Successful Response (`200 OK`):**

```json
{
  "balance": 0  // integer number of units
}
```

**Errors:**

| Status | Condition                    | Response           |
|-------:|------------------------------|--------------------|
| 400    | Invalid address              | Plain text message |

### `GET /wallets/{address}/transactions`

Returns all on‑chain transactions where the specified address appears in at least one output.  Transactions are returned in their full form.

**Response (`200 OK`):** an array of transaction objects.  Each transaction has the following structure (byte slices are Base64‑encoded by Go’s JSON encoder):

```json
[
  {
    "ID": "string",             // Base64‑encoded transaction ID
    "Vin": [                     // array of inputs
      {
        "Txid": "string",      // Base64‑encoded referenced txid
        "Vout": 0,              // index of the referenced output
        "Signature": "string",  // Base64‑encoded ECDSA signature
        "PubKey": "string"      // Base64‑encoded public key
      }
    ],
    "Vout": [                    // array of outputs
      {
        "Value": 0,             // integer amount
        "PubKeyHash": "string"  // Base64‑encoded pubKeyHash
      }
    ]
  }
]
```

**Errors:**

| Status | Condition                                       | Response           |
|-------:|-------------------------------------------------|--------------------|
| 400    | Invalid address or decoding error               | Plain text message |

## Transactions

### `POST /transactions`

Submits a new transaction to transfer funds between wallets.  The transaction is constructed and signed server‑side, mined into a new block immediately and the UTXO set is rebuilt.  The private key must correspond to the `from` address.

**Request Body:**

```json
{
  "from": "string",     // sender wallet address (hex)
  "to": "string",       // receiver wallet address (hex)
  "amount": 0,           // positive integer amount to send
  "privKey": "string"   // hex‑encoded private key of sender (D value)
}
```

**Successful Response (`200 OK`):**

```json
{
  "status": "transaction mined"
}
```

**Errors:**

| Status | Condition                                                        | Response           |
|-------:|------------------------------------------------------------------|--------------------|
| 400    | Malformed JSON                                                   | Plain text message |
| 400    | `from` or `to` address fails validation                          | Plain text message |
| 400    | `amount` is zero or negative                                    | Plain text message |
| 400    | Private key cannot be decoded                                    | Plain text message |
| 400    | Insufficient unspent outputs to cover the requested amount        | Plain text message |
| 400    | Transaction creation or signature verification fails             | Plain text message |

## Block Explorer

### `GET /blocks`

Returns a summary of every block in the chain.  Blocks are ordered by height (genesis at index 0).

**Successful Response (`200 OK`):** an array of block summaries:

```json
[
  {
    "index": 0,          // height of the block
    "timestamp": 0,      // UNIX timestamp
    "hash": "string",    // hex‑encoded block hash
    "prev_hash": "string",// hex‑encoded previous block hash
    "tx_count": 1        // number of transactions in the block
  }
]
```

### `GET /blocks/{index}`

Returns the full details of a block by its index (height).

**Path Parameters:**

| Name  | Type | Description                        |
|-------|------|------------------------------------|
| index | int  | Zero‑based block height to retrieve |

**Successful Response (`200 OK`):**

```json
{
  "Timestamp": 0,
  "Transactions": [ /* array of transactions (see above) */ ],
  "PrevHash": "string", // Base64‑encoded bytes of previous hash
  "Hash": "string",     // Base64‑encoded bytes of the block hash
  "Nonce": 0            // integer nonce produced by proof‑of‑work
}
```

**Errors:**

| Status | Condition                     | Response           |
|-------:|-------------------------------|--------------------|
| 400    | `index` is not a valid number | Plain text message |
| 404    | No block exists at that index | Plain text message |

## Wallet Reporting

### `GET /reports/wallet/{address}`

Generates a report on a specific wallet by aggregating transaction history and zakat deductions from the database.  Supabase must be configured or the endpoint will return a server error.

**Path Parameters:**

| Name    | Type   | Description                              |
|---------|--------|------------------------------------------|
| address | string | Wallet address (hex‑encoded public hash) |

**Successful Response (`200 OK`):**

```json
{
  "wallet_address": "string",
  "balance": 0,
  "total_sent": 0,
  "total_received": 0,
  "total_zakat": 0,
  "transactions": [ /* array of transaction records */ ],
  "zakat_records": [ /* array of zakat records */ ]
}
```

Each transaction record includes `txid`, `block_hash`, `sender`, `receiver`, `amount`, `timestamp`, `type` and a `raw_json` object containing the full serialized transaction.  Each zakat record includes `id`, `user_id`, `wallet_address`, `amount`, `block_hash` and `created_at` (ISO 8601 timestamp).

**Errors:**

| Status | Condition                                              | Response           |
|-------:|--------------------------------------------------------|--------------------|
| 400    | Empty or invalid address                               | Plain text message |
| 500    | Database not configured or retrieval failure           | Plain text message |

## System Logs

### `GET /logs/system`

Returns the most recent system log entries from the database.  Requires Supabase configuration.

**Query Parameters:**

| Name  | Type | Description                                                       | Default |
|-------|------|-------------------------------------------------------------------|---------|
| limit | int  | Maximum number of log entries to return (1 – 1000)               | 100     |

**Successful Response (`200 OK`):**

```json
{
  "logs": [
    {
      "id": "string",
      "level": "string",      // info, warn or error
      "type": "string",       // context of the event (e.g. otp_generated)
      "message": "string",    // human‑readable message
      "ip": "string",         // client IP address
      "timestamp": "string"    // ISO 8601 timestamp
    }
  ]
}
```

**Errors:**

| Status | Condition                          | Response           |
|-------:|------------------------------------|--------------------|
| 500    | Database not configured or failure | Plain text message |

## Zakat Deduction

### `POST /zakat/run`

Calculates and deducts Zakat (2.5%) from every wallet profile in the database.  For each eligible wallet, the server builds and mines a transaction sending the computed amount to the Zakat pool wallet (`ZAKAT_WALLET_ADDRESS`), persists the block, transaction and zakat record, rebuilds the UTXO set and logs the event.  This endpoint is typically restricted to administrators.

**Request Body:** *none*

**Successful Response (`200 OK`):**

```json
{
  "total_wallets": 0,    // total number of wallet profiles scanned
  "processed": 0,        // number of wallets from which zakat was deducted
  "total_zakat": 0,      // total units deducted across all wallets
  "block_hashes": [ "string" ] // array of mined block hashes (hex)
}
```

**Errors:**

| Status | Condition                                              | Response           |
|-------:|--------------------------------------------------------|--------------------|
| 500    | Database not configured                                | Plain text message |
| 500    | `ZAKAT_WALLET_ADDRESS` env var not set                 | Plain text message |
| 500    | Failure while listing wallet profiles or persisting data | Plain text message |

## Admin Faucet

### `POST /admin/fund`

Creates a coinbase transaction that credits the specified wallet address.  Intended to serve as a faucet for development and demonstration.  Note that the coinbase transaction always mints a fixed reward defined in the blockchain layer (15 000 units) regardless of the `amount` field in the request; however the `amount` is stored with the transaction in the database.

**Request Body:**

```json
{
  "address": "string",  // recipient wallet address (hex)
  "amount": 0            // positive integer (recorded only)
}
```

**Successful Response (`200 OK`):**

```json
{
  "address": "string",
  "amount": 0,
  "block_hash": "string" // hex‑encoded hash of the newly mined block
}
```

**Errors:**

| Status | Condition                                 | Response           |
|-------:|-------------------------------------------|--------------------|
| 400    | Invalid JSON, empty address or amount ≤ 0 | Plain text message |
| 400    | Wallet address fails validation            | Plain text message |
