package cli

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/aspertheghost/blockchain/internal/blockchain"
	"github.com/aspertheghost/blockchain/internal/proofofwork"
	"github.com/aspertheghost/blockchain/internal/transaction"
	"github.com/aspertheghost/blockchain/internal/wallet"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
	bc := blockchain.CreateBlockchain(address)
	bc.Close()
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := blockchain.NewBlockchain(address)
	defer bc.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  createwallet -name NAME - Create a new wallet with optional name")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printChain() {
	bc := blockchain.NewBlockchain("ANY_ADDRESS")
	defer bc.Close()

	bci := bc.Iterator()

	for {
		blk := bci.Next()

		fmt.Printf("Prev. hash: %x\n", blk.PrevBlockHash)
		fmt.Printf("Hash: %x\n", blk.Hash)
		pow := proofofwork.NewProofOfWork(blk)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(blk.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	// Resolve names to addresses if they are aliases
	resolvedFrom := resolveAddress(from)
	resolvedTo := resolveAddress(to)

	bc := blockchain.NewBlockchain(resolvedFrom)
	defer bc.Close()

	// Load wallets to get the private key
	ws, err := wallet.NewWallets()
	if err != nil {
		log.Printf("ERROR: Failed to load wallets: %v\n", err)
		return
	}

	// Get the sender's wallet
	senderWallet := ws.GetWallet(resolvedFrom)
	if senderWallet == nil {
		fmt.Printf("ERROR: No wallet found for address '%s'\n", resolvedFrom)
		return
	}

	// First check balance before creating transaction
	balance := 0
	UTXOs := bc.FindUTXO(resolvedFrom)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("DEBUG: Checking balance for '%s': found %d UTXOs, total balance: %d\n", resolvedFrom, len(UTXOs), balance)

	if balance < amount {
		fmt.Printf("ERROR: Not enough funds. Address '%s' has balance %d, but tried to send %d.\n", resolvedFrom, balance, amount)
		return
	}

	fmt.Printf("DEBUG: Creating transaction from %s to %s for %d\n", resolvedFrom, resolvedTo, amount)

	// Reconstruct the private key from stored bytes
	privKey := reconstructPrivateKey(senderWallet.PrivateKey)

	tx, err := transaction.NewUTXOTransaction(resolvedFrom, resolvedTo, amount, bc, privKey, senderWallet.PublicKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to create transaction: %v\n", err)
		return
	}

	fmt.Printf("DEBUG: Transaction created, mining block...\n")
	bc.MineBlock([]*transaction.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CLI) createWallet(name string) {
	ws, err := wallet.NewWallets()
	if err != nil {
		log.Printf("ERROR: Failed to create wallets: %v\n", err)
		return
	}

	ws.CreateWalletWithName(name)
}

// resolveAddress resolves a name to an address if it exists in aliases
func resolveAddress(name string) string {
	aliases, err := wallet.NewWalletAliases()
	if err == nil {
		if addr := aliases.GetAddress(name); addr != "" {
			return addr
		}
	}
	return name
}

// reconstructPrivateKey
func reconstructPrivateKey(privateKeyBytes []byte) *ecdsa.PrivateKey {
	// The private key is stored as D.Bytes()
	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
	}

	priv.D = new(big.Int).SetBytes(privateKeyBytes)

	// Recompute public key from private key
	priv.PublicKey.X, priv.PublicKey.Y = priv.Curve.ScalarBaseMult(privateKeyBytes)

	return priv
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address or name")
	sendTo := sendCmd.String("to", "", "Destination wallet address or name")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	createWalletName := createWalletCmd.String("name", "", "Optional name for the wallet")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		// Resolve name if it's an alias
		address := resolveAddress(*getBalanceAddress)
		cli.getBalance(address)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(*createWalletName)
	}
}
