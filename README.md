# Go Blockchain Implementation

A secure blockchain implementation written in Go that demonstrates core blockchain concepts including blocks, proof-of-work consensus, transactions with ECDSA digital signatures, and the UTXO (Unspent Transaction Output) model.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Examples](#examples)
- [Understanding Blockchain](#understanding-blockchain)
- [Architecture](#architecture)
- [Technical Details](#technical-details)
- [Wallet System](#wallet-system)

## Quick Start

```bash
# Build the blockchain binary
make build

# Create a new wallet (automatically creates blockchain and funds with 10 coins)
./blockchain createwallet -name "alice"

# Check your balance
./blockchain getbalance -address "alice"

# Send some coins
./blockchain send -from "alice" -to "bob" -amount 5

# Create another wallet to receive coins
./blockchain createwallet -name "bob"

# Send coins using aliases
./blockchain send -from "alice" -to "bob" -amount 5

# Print the entire blockchain
./blockchain printchain
```

## Installation

### Prerequisites

- Go 1.16 or higher
- Make

### Setup

1. Clone the repository
2. Install dependencies:

```bash
make deps
```

3. Build the binary:

```bash
make build
```

Or use the install target to install globally:

```bash
make install
```

## Usage

This blockchain uses an ECDSA-based wallet system for secure key management and transaction signing.

### Recommended Usage Flow

1. **Create a wallet** using `createwallet -name "yourname"` command (automatically creates blockchain and funds with 10 coins)
2. **Create additional wallets** for sending/receiving
3. **Check balance** and **send transactions** using wallet names or addresses

### About Addresses and Aliases

- **Addresses** are base58-encoded strings derived from ECDSA public keys
- **Aliases** are human-readable names you can assign to addresses for easier use
- You can use either the address or the alias in any command

## Commands

### createwallet

Creates a new ECDSA wallet and generates a secure address. The wallet is automatically funded with 10 coins.

```bash
./blockchain createwallet
```

Or with an optional name for easy reference:

```bash
./blockchain createwallet -name "mywallet"
```

This command:
1. Generates a new ECDSA key pair (P-256 curve)
2. Derives a base58-encoded address from the public key
3. Saves the wallet to `wallet.dat`
4. If the blockchain doesn't exist, creates it with this wallet as the genesis recipient
5. Funds the wallet with 10 coins (genesis block or coinbase transaction)
6. If a name is provided, creates a name-to-address mapping in `wallet-aliases.dat`

### createblockchain

Creates a new blockchain and sends the genesis block reward to the specified address.

```bash
./blockchain createblockchain -address "your-address"
```

This command:
1. Creates a new database file (blockchain.db)
2. Mines the genesis block with a coinbase transaction
3. Sends the initial subsidy (10 coins) to your address

**Note:** Usually you don't need this command since `createwallet` automatically creates the blockchain.

### getbalance

Retrieves the balance for a specific address by scanning the blockchain for unspent transaction outputs (UTXOs).

```bash
./blockchain getbalance -address "your-address"
```

Or use a wallet name:

```bash
./blockchain getbalance -address "alice"
```

### send

Sends coins from one address to another. The system finds unspent outputs from the sender, spends them, and creates new outputs for the recipient (and change back to sender if applicable).

```bash
./blockchain send -from "sender-address-or-name" -to "recipient-address-or-name" -amount 10
```

This command:
1. Loads the sender's wallet from wallet.dat
2. Reconstructs the private key from stored bytes
3. Creates and signs a transaction with ECDSA
4. Mines a new block containing the transaction
5. Validates the transaction signatures

### printchain

Prints all blocks in the blockchain from the genesis block to the latest block. Shows block hashes, previous hashes, and proof-of-work validation.

```bash
./blockchain printchain
```

## Examples

### Example 1: Creating Your First Wallet and Blockchain

```bash
# Build the project
make build

# Create a new wallet with a name
./blockchain createwallet -name "alice"
```

Output:
```
Wallet 'alice' created with address: 1H1S3W8d3ZPx4Y9s2F7k6m5n8p1q4r7t
Created new blockchain with genesis block reward of 10 coins for alice
```

```bash
# Check your balance (should be 10 - the genesis subsidy)
./blockchain getbalance -address "alice"
```

Output:
```
Balance of 'alice': 10
```

### Example 2: Creating Additional Wallets and Sending Transactions

```bash
# Create a second wallet
./blockchain createwallet -name "bob"
```

Output:
```
Wallet 'bob' created with address: 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P
Added 10 coins to wallet bob
```

```bash
# Send 5 coins from alice to bob using names
./blockchain send -from "alice" -to "bob" -amount 5
```

Output:
```
Success!
```

```bash
# Check balances
./blockchain getbalance -address "alice"
./blockchain getbalance -address "bob"
```

Output:
```
Balance of 'alice': 5
Balance of 'bob': 15
```

You can still use addresses directly if preferred:
```bash
./blockchain send -from "1H1S3W8d3ZPx4Y9s2F7k6m5n8p1q4r7t" -to "1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P" -amount 5
```

### Example 3: Multiple Transactions and Change

```bash
# Create additional wallets with names
./blockchain createwallet -name "charlie"
./blockchain createwallet -name "david"

# Send 3 coins from charlie to david
./blockchain send -from "charlie" -to "david" -amount 3
```

Note: Charlie had 10 coins initially. After sending 3 to David:
- David has 10 + 3 = 13 coins
- Charlie has 10 - 3 = 7 coins (change is automatically sent back)

```bash
# Check all balances
./blockchain getbalance -address "charlie"
./blockchain getbalance -address "david"
```

Output:
```
Balance of 'charlie': 7
Balance of 'david': 13
```

### Example 4: Viewing the Blockchain

```bash
./blockchain printchain
```

Sample output:
```
Prev. hash: 0000000000000000000000000000000000000000000000000000000000000000
Hash: 00a3c9c58c2a7a2e4b9d8c7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6
PoW: true

Prev. hash: 00a3c9c58c2a7a2e4b9d8c7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6
Hash: 0058f9e8a7d6c5b4a3928172636450f1e2d3c4b5a69788907f6e5d4c3b2a1900f
PoW: true

...
```

### Example 5: Handling Insufficient Funds

```bash
# Try to send more than available balance
./blockchain send -from "alice" -to "bob" -amount 100
```

Output:
```
ERROR: Not enough funds. Address 'alice' has balance 5, but tried to send 100.
```

---

# Understanding Blockchain

This section provides educational content explaining blockchain concepts.

## What is a Blockchain?

A blockchain is a distributed, decentralized digital ledger that records transactions across many computers. The key characteristics are:

- **Decentralization**: No single entity controls the entire network
- **Immutability**: Once data is written, it cannot be changed
- **Transparency**: All transactions are visible to participants
- **Security**: Cryptographic hashing ensures data integrity

Think of a blockchain as a shared spreadsheet that thousands of people have identical copies of. When someone makes a change, everyone updates their copy. This makes it extremely difficult to cheat because you would need to change everyone's spreadsheet simultaneously.

## Blocks

### What is a Block?

A block is a container data structure that holds a batch of transactions. Each block contains:

1. **Block Header**: Metadata about the block
   - Timestamp: When the block was created
   - Previous Block Hash: Hash of the previous block (creates the chain)
   - Merkle Root: Hash of all transactions in the block
   - Nonce: Number used in proof-of-work
   - Block Hash: Hash of the current block header

2. **Block Body**: The actual transaction data

### Block Structure in This Implementation

```go
type Block struct {
    Timestamp     int64                          // Unix timestamp
    Transactions  []*transaction.Transaction      // List of transactions
    PrevBlockHash []byte                         // Hash of previous block
    Hash          []byte                         // Current block hash
    Nonce         int                            // Proof-of-work nonce
}
```

### The Genesis Block

The genesis block is the first block in a blockchain. It is hardcoded and serves as the foundation. In this implementation:

- Created when `createblockchain` is first run or when first wallet is created
- Contains a coinbase transaction with the message: "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks" (a reference to Bitcoin's genesis block)
- Receives the initial subsidy of 10 coins

### Block Chaining

Blocks are linked together through cryptographic hashes. Each block contains the hash of the previous block, creating an unbroken chain. This provides:

1. **Tamper Evidence**: Changing any transaction in a block changes its hash, breaking the link to the next block
2. **Sequential Order**: Blocks must be processed in order
3. **History Preservation**: All historical transactions are preserved in the chain

## Transactions

### What is a Transaction?

A transaction is a record of value transfer between addresses. Each transaction has:

- **Inputs (Vin)**: References to previous outputs being spent
- **Outputs (Vout)**: New outputs created, specifying recipient addresses and amounts

### The UTXO Model

This implementation uses the Unspent Transaction Output (UTXO) model, the same model used by Bitcoin.

#### How UTXO Works:

1. **Outputs are Created**: When coins are sent, new outputs are created with specific values and recipient addresses

2. **Outputs are Spent**: An output can only be spent once, by referencing it as an input in a new transaction

3. **Unspent Outputs = Balance**: Your balance is the sum of all unspent outputs locked to your address

#### Transaction Structure:

```go
type Transaction struct {
    ID   []byte      // Transaction ID (hash of transaction)
    Vin  []TXInput   // List of inputs
    Vout []TXOutput  // List of outputs
}

type TXInput struct {
    Txid      []byte    // Previous transaction ID
    Vout      int       // Output index in that transaction
    Signature []byte    // ECDSA signature (R || S bytes)
    PubKey    []byte    // Public key for signature verification
}

type TXOutput struct {
    Value        int    // Amount of coins
    ScriptPubKey string // Locking script (recipient address)
}
```

#### Digital Signatures

Each transaction input now includes a digital signature created with ECDSA:

- **Signature**: Concatenated R and S bytes from ECDSA signature
- **Public Key**: Full public key (X || Y coordinates) for verification
- **Verification**: The signature proves the sender owns the private key corresponding to the public key

#### Transaction Flow Example:

1. Alice has 10 coins (one UTXO of value 10 locked to "Alice")
2. Alice sends 5 coins to Bob:
   - Input: Spend Alice's 10-coin output
   - Output 1: 5 coins locked to "Bob"
   - Output 2: 5 coins (change) locked to "Alice"
3. Alice now has one UTXO of 5 coins, Bob has one UTXO of 5 coins

### Coinbase Transactions

A coinbase transaction is the first transaction in a block, created by the miner as a reward for mining the block. It has:

- No inputs (or inputs with zero transaction ID and output index -1)
- One output with the mining reward
- Includes arbitrary data (like the genesis message)

### Transaction Validation

Transactions are validated by:
1. Checking that referenced inputs exist and are unspent
2. Verifying ECDSA signatures on each input
3. Ensuring that total outputs do not exceed total inputs
4. Confirming the sender's public key matches the address

#### Signature Verification

Each transaction input signature is verified using ECDSA:

```go
func (tx *Transaction) Verify() bool {
    txCopy := tx.TrimmedCopy()
    
    for i, vin := range tx.Vin {
        dataToVerify := txCopy.getDataToSign(i)
        
        // Parse signature R and S values
        r := new(big.Int)
        s := new(big.Int)
        sigLen := len(signature) / 2
        r.SetBytes(signature[:sigLen])
        s.SetBytes(signature[sigLen:])
        
        // Reconstruct public key from X and Y coordinates
        x := new(big.Int).SetBytes(pubKey[:len(pubKey)/2])
        y := new(big.Int).SetBytes(pubKey[len(pubKey)/2:])
        
        ecdsaPubKey := &ecdsa.PublicKey{
            Curve: elliptic.P256(),
            X:     x,
            Y:     y,
        }
        
        if !ecdsa.Verify(ecdsaPubKey, dataToVerify, r, s) {
            return false
        }
    }
    return true
}
```

## Proof-of-Work

### What is Proof-of-Work?

Proof-of-Work (PoW) is a consensus mechanism that requires computational work to add new blocks. Miners compete to solve a mathematical puzzle, and the first to solve it gets to add the next block.

### How Proof-of-Work Works in This Implementation:

1. **Target**: A 256-bit number. The hash must be less than this target to be valid.

2. **Difficulty**: This implementation uses 16 leading zero bits (targetBits = 16).

3. **Mining Process**:
   - The miner increments a nonce value
   - For each nonce, the block data is hashed (SHA-256)
   - If the hash is less than the target, mining is successful
   - Otherwise, try the next nonce

4. **Code Example**:

```go
func (pow *ProofOfWork) Run() (int, []byte) {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0

    for nonce < maxNonce {
        data := pow.prepareData(nonce)
        hash = sha256.Sum256(data)
        hashInt.SetBytes(hash[:])

        if hashInt.Cmp(pow.target) == -1 {
            break  // Found valid hash!
        }
        nonce++
    }
    return nonce, hash[:]
}
```

### Why Proof-of-Work?

- **Security**: Attacking the network requires more than 50% of total mining power
- **Fairness**: Anyone can participate; only computational power matters
- **Immutability**: Changing historical blocks would require re-mining all subsequent blocks

### Difficulty Adjustment

In production blockchains like Bitcoin, difficulty adjusts dynamically to maintain consistent block times. This implementation uses a fixed difficulty (16 bits) for simplicity and faster testing.

## Mining

### What is Mining?

Mining is the process of:
1. Collecting pending transactions from the network
2. Creating a new block with these transactions
3. Solving the proof-of-work puzzle
4. Broadcasting the valid block

### Mining in This Implementation:

```bash
./blockchain send -from "Alice" -to "Bob" -amount 5
```

When a transaction is sent:
1. A new transaction is created and signed with the sender's private key
2. The blockchain finds unspent outputs from the sender
3. A new block is mined with this transaction
4. The block is added to the chain

### Mining Rewards

Miners receive:
- **Subsidy**: Newly created coins (10 in this implementation)
- **Transaction Fees**: Fees from included transactions

In this simplified version, the sender address receives the mining reward (through the coinbase transaction).

---

# Architecture

## Project Structure

```
blockchain/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── block/
│   │   └── block.go           # Block structure and methods
│   ├── blockchain/
│   │   └── blockchain.go      # Blockchain implementation
│   ├── cli/
│   │   └── cli.go             # Command-line interface with wallet support
│   ├── proofofwork/
│   │   └── proofofwork.go     # Proof-of-work consensus
│   ├── transaction/
│   │   └── transaction.go     # Transaction handling with ECDSA signatures
│   ├── utils/
│   │   └── utils.go           # Utility functions
│   └── wallet/
│       ├── wallet.go          # ECDSA wallet and address generation
│       └── aliases.go         # Wallet aliases for easy reference
├── blockchain.db              # Database file (created on first run)
├── wallet.dat                 # Wallet file (created when wallets are used)
├── wallet-aliases.dat         # Wallet aliases file
├── go.mod
├── go.sum
└── Makefile
```

## Database Schema

This implementation uses BoltDB, a simple key-value store. The database contains one bucket:

**Bucket: "blocks"**

| Key | Value |
|-----|-------|
| Block Hash (bytes) | Serialized Block |
| "l" (last indicator) | Latest block hash |

## Files

The blockchain implementation uses the following files for persistence:

| File | Description |
|------|-------------|
| `blockchain.db` | BoltDB database storing all blocks and transactions |
| `wallet.dat` | Serialized wallets containing ECDSA key pairs |
| `wallet-aliases.dat` | Name-to-address mappings for wallet aliases |

---

# Technical Details

## Cryptography

### Hash Functions

The implementation uses SHA-256 for:
- Block hashing
- Transaction ID generation
- Public key hashing
- Address checksum generation

### Digital Signatures (ECDSA)

This implementation uses Elliptic Curve Digital Signature Algorithm (ECDSA) with the P-256 curve for transaction signing:

- **Key Generation**: Uses `crypto/ecdsa` with `elliptic.P256()` curve
- **Signing**: Each transaction input is signed using `ecdsa.Sign()`
- **Verification**: Signatures are verified using `ecdsa.Verify()`

### Signature Structure

```go
type TXInput struct {
    Txid      []byte    // Previous transaction ID
    Vout      int       // Output index in that transaction
    Signature []byte    // ECDSA signature (R || S bytes)
    PubKey    []byte    // Public key for verification
}
```

### Signature Process

1. **Create trimmed copy** of transaction (without signatures)
2. **Hash the transaction** using SHA-256
3. **Sign with private key**: `ecdsa.Sign(rand.Reader, privKey, hash)`
4. **Store signature** concatenated R || S bytes

```go
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey) {
    txCopy := tx.TrimmedCopy()
    for i := range txCopy.Vin {
        txCopy.Vin[i].Signature = nil
        txCopy.Vin[i].PubKey = nil

        dataToSign := txCopy.getDataToSign(i)
        r, s, err := ecdsa.Sign(rand.Reader, privKey, dataToSign)
        if err != nil {
            log.Panic(err)
        }

        signature := append(r.Bytes(), s.Bytes()...)
        tx.Vin[i].Signature = signature
    }
}
```

### Wallet Key Storage

Private keys are stored as raw bytes in wallet.dat:

1. **Private Key Storage**: The D value (scalar) of the ECDSA private key is stored as bytes
2. **Public Key Storage**: The X and Y coordinates are concatenated and stored
3. **Key Reconstruction**: Private and public keys are reconstructed from stored bytes when needed

```go
// Storage in wallet.dat
type Wallet struct {
    PrivateKey []byte  // D value bytes
    PublicKey  []byte  // X || Y coordinates
}

// Reconstruction from bytes
func reconstructPrivateKey(privateKeyBytes []byte) *ecdsa.PrivateKey {
    priv := &ecdsa.PrivateKey{
        PublicKey: ecdsa.PublicKey{
            Curve: elliptic.P256(),
        },
    }
    priv.D = new(big.Int).SetBytes(privateKeyBytes)
    priv.PublicKey.X, priv.PublicKey.Y = priv.Curve.ScalarBaseMult(privateKeyBytes)
    return priv
}
```

### Hash Structure

Block hash is computed from:
- Previous block hash
- Transaction Merkle root (hash of all transaction IDs)
- Timestamp
- Target bits
- Nonce

```go
func (pow *ProofOfWork) prepareData(nonce int) []byte {
    return bytes.Join([][]byte{
        pow.block.PrevBlockHash,
        pow.block.HashTransactions(),
        utils.IntToHex(pow.block.Timestamp),
        utils.IntToHex(int64(targetBits)),
        utils.IntToHex(int64(nonce)),
    }, []byte{})
}
```

## Serialization

Blocks and transactions are serialized using Go's `encoding/gob` package for storage in BoltDB.

### Block Serialization:

```go
func (b *Block) Serialize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    err := encoder.Encode(b)
    // ...
    return result.Bytes()
}
```

## Limitations and Extensions

This implementation is educational and has several simplifications:

### Current Features (Completed):

1. ✅ **Cryptographic Wallets**: ECDSA-based wallet system with P-256 curve
2. ✅ **Digital Signatures**: Transaction signing and verification
3. ✅ **Secure Addresses**: Base58-encoded addresses from public keys
4. ✅ **Persistent Storage**: Wallets saved to wallet.dat file
5. ✅ **Wallet Management**: Create and manage multiple wallets
6. ✅ **Wallet Aliases**: Create memorable names for wallet addresses
7. ✅ **Auto-funding**: New wallets automatically receive 10 coins
8. ✅ **Auto-blockchain Creation**: Blockchain is created automatically with first wallet

### Current Limitations:

1. **No Network**: Runs as a single node
2. **No Mempool**: Transactions are mined immediately
3. **Fixed Difficulty**: No dynamic difficulty adjustment
4. **No Transaction Fees**: Simplified subsidy model
5. **No Merkle Tree**: Uses simple transaction hash concatenation

### Possible Extensions:

1. **Add P2P Network**: Enable communication between nodes
2. **Mempool**: Buffer unconfirmed transactions
3. **Difficulty Adjustment**: Implement difficulty retargeting
4. **Transaction Fees**: Add fee calculation and collection
5. **Merkle Trees**: Implement proper Merkle tree for efficient validation
6. **SPV Support**: Add Simplified Payment Verification
7. **Encrypted Wallets**: Add password protection for wallet.dat

## Further Learning

To extend your knowledge of blockchain technology:

1. **Bitcoin Whitepaper**: Read Satoshi Nakamoto's original Bitcoin whitepaper
2. **Ethereum**: Study the Ethereum virtual machine and smart contracts
3. **Consensus Mechanisms**: Explore Proof-of-Stake, Delegated Proof-of-Stake
4. **Cryptography**: Learn about elliptic curve cryptography and digital signatures
5. **Smart Contracts**: Understand Turing-complete contract execution

## Wallet System

This implementation includes a complete ECDSA-based wallet system for secure key management and transaction signing.

### Wallet Architecture

The wallet system provides:

- **Cryptographic Security**: Uses ECDSA with P-256 curve for digital signatures
- **Address Generation**: Creates secure base58-encoded addresses from public keys
- **Persistent Storage**: Wallets are saved to `wallet.dat` file
- **Transaction Signing**: Each transaction input is signed with the sender's private key
- **Wallet Aliases**: Create memorable names for wallet addresses for easier use
- **Auto-funding**: New wallets are automatically funded with 10 coins

### Wallet Aliases

You can assign memorable names to wallet addresses for easier reference in commands:

```bash
# Create a wallet with a name
./blockchain createwallet -name "alice"

# Use the name in other commands
./blockchain getbalance -address "alice"
./blockchain send -from "alice" -to "bob" -amount 5
```

### Auto-funding New Wallets

When you create a new wallet:

1. If no blockchain exists, it creates one with your wallet as the genesis recipient
2. Your wallet receives 10 coins from the genesis block (or a coinbase transaction)
3. You can immediately start sending transactions

This makes the initial setup much simpler - just create a wallet and you're ready to go!

### Wallet Aliases Management

Wallet aliases are stored in `wallet-aliases.dat` and support the following operations:

```go
// Create a new aliases manager
aliases, err := wallet.NewWalletAliases()

// Set a name-to-address mapping
aliases.SetAlias("alice", "1H1S3W8d3ZPx4Y9s2F7k6m5n8p1q4r7t")

// Get address from name
address := aliases.GetAddress("alice") // Returns "1H1S3W8d3ZPx4Y9s2F7k6m5n8p1q4r7t"

// Get name from address
name := aliases.GetAlias("1H1S3W8d3ZPx4Y9s2F7k6m5n8p1q4r7t") // Returns "alice"

// Save to file
aliases.SaveToFile()

// Load from file (automatic on NewWalletAliases)
aliases.LoadFromFile()

// Print all aliases
aliases.PrintAliases()
```

### Wallet Structure

```go
type Wallet struct {
    PrivateKey []byte  // ECDSA private key bytes (D value)
    PublicKey  []byte  // ECDSA public key bytes (X || Y coordinates)
}

type Wallets struct {
    Wallets map[string]*Wallet  // Maps addresses to wallets
}
```

### Address Generation

Addresses are generated from public keys using the following process:

1. **Hash the public key** using SHA-256
2. **Add version byte** (0x00 for main network)
3. **Calculate checksum** (SHA-256 of versioned hash, first 4 bytes)
4. **Encode to base58** for human-readable format

```go
func (w *Wallet) GetAddress() string {
    pubKeyHash := HashPubKey(w.PublicKey)

    versionedPayload := append([]byte{byte(0x00)}, pubKeyHash...)
    checksum := checksum(versionedPayload)

    fullPayload := append(versionedPayload, checksum...)
    address := base58Encode(fullPayload)

    return string(address)
}
```

### Using Wallets

```go
// Create a new wallet
wallet := wallet.NewWallet()
address := wallet.GetAddress()

// Create a wallets collection and add wallet
wallets, _ := wallet.NewWallets()
address := wallets.CreateWallet()

// Save wallets to file
wallets.SaveToFile()
```

### Transaction Signing

Each transaction input requires a digital signature created with the sender's private key:

1. **Load wallet** from wallet.dat using address
2. **Reconstruct private key** from stored bytes
3. **Create transaction copy** (trimmed copy without signatures)
4. **Sign each input** using ECDSA.Sign with private key
5. **Store signature** in the TXInput

```go
// Reconstruct private key from bytes
privKey := reconstructPrivateKey(senderWallet.PrivateKey)

// Sign transaction with private key
tx.Sign(privKey)

// Verify signatures
valid := tx.Verify()
```

### Security Features

- **Private Key Storage**: Private keys stored as raw bytes in wallet.dat
- **Signature Verification**: All transactions verify signatures before mining
- **Address Validation**: Checksums ensure address integrity
- **No Plaintext Keys**: Raw private keys are never exposed in transaction data

---

## Makefile Targets

The project includes a Makefile for common operations:

```bash
make all           # Download dependencies and build (default)
make deps          # Download Go dependencies
make build         # Build the binary
make run           # Build and run the program
make clean         # Remove binary, databases, and wallet files
make test          # Run all tests
make install       # Install binary to GOPATH/bin
make fmt           # Format Go code
make vet           # Run go vet
make help          # Show available targets
```

## License

This project is for educational purposes.

