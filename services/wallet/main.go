package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	// wallet "test/wallet"
	"time"

	wallet "github.com/Shivam-Patel-G/blackhole-blockchain/services/wallet/wallet"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 🧩 MongoDB setup
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println("MongoDB connection error:", err)
		return
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Println("MongoDB ping failed:", err)
		return
	}

	db := client.Database("walletdb") // Use your DB name
	wallet.UserCollection = db.Collection("users")
	wallet.WalletCollection = db.Collection("wallets")

	// Initialize blockchain client
	if err := wallet.InitBlockchainClient(4000); err != nil { // Use different port for wallet
		log.Fatalf("Failed to initialize blockchain client: %v", err)
	}

	// Connect to blockchain node (assuming it's running on port 3000)
	blockchainPeerAddr := "/ip4/25.2.105.97/tcp/3000/p2p/12D3KooWLNz4GcCqkKcsjWrcV5RixrrBneeEByAKULyoTN9fevxy" // Replace with actual peer ID
	fmt.Println("⚠️ Please start the blockchain node first and update the peer address above")
	fmt.Printf("Attempting to connect to blockchain node: %s\n", blockchainPeerAddr)

	// Try to connect to blockchain node (this will fail if node is not running)
	if err := wallet.DefaultBlockchainClient.ConnectToBlockchain(blockchainPeerAddr); err != nil {
		fmt.Printf("⚠️ Failed to connect to blockchain node: %v\n", err)
		fmt.Println("⚠️ Wallet will work in offline mode. Start blockchain node and restart wallet for full functionality.")
	}

	fmt.Println("Welcome to the Wallet CLI")
	var loggedInUser *wallet.User = nil

	for {

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel() // make sure to cancel context after the operation

		fmt.Println("\nChoose an option:")
		if loggedInUser == nil {
			fmt.Println("1. Register")
			fmt.Println("2. Login")
			fmt.Println("3. Exit")
		} else {
			fmt.Printf("Logged in as: %s\n", loggedInUser.Username)
			fmt.Println("1. Generate Wallet from Mnemonic")
			fmt.Println("2. Logout")
			fmt.Println("3. Show my wallets")
			fmt.Println("4. Show My Wallet Details")
			fmt.Println("5. Exit")
			fmt.Println("6. Check Token Balance")
			fmt.Println("7. Transfer Tokens")
			fmt.Println("8. Stake Tokens")
		}

		fmt.Print("Enter your choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if loggedInUser == nil {
			switch choice {
			case "1":
				fmt.Print("Enter username: ")
				username, _ := reader.ReadString('\n')
				username = strings.TrimSpace(username)

				fmt.Print("Enter password: ")
				password, _ := reader.ReadString('\n')
				password = strings.TrimSpace(password)

				user, err := wallet.RegisterUser(ctx, username, password)
				if err != nil {
					fmt.Println("Registration failed:", err)
				} else {
					fmt.Println("Registered successfully!")
					fmt.Printf("User ID: %v\n", user.ID)
				}

			case "2":
				fmt.Print("Enter username: ")
				username, _ := reader.ReadString('\n')
				username = strings.TrimSpace(username)

				fmt.Print("Enter password: ")
				password, _ := reader.ReadString('\n')
				password = strings.TrimSpace(password)

				user, err := wallet.AuthenticateUser(ctx, username, password)
				if err != nil {
					fmt.Println("Login failed:", err)
				} else {
					fmt.Println("Login successful!")
					fmt.Printf("Welcome, %s (User ID: %v)\n", user.Username, user.ID)
					loggedInUser = user
				}

			case "3":
				fmt.Println("Exiting Wallet CLI.")
				return

			default:
				fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
			}
		} else {
			switch choice {
			case "1":
				fmt.Print("Enter wallet name: ")
				walletName, _ := reader.ReadString('\n')
				walletName = strings.TrimSpace(walletName)

				fmt.Print("Enter your password to secure the wallet: ")
				password, _ := reader.ReadString('\n')
				password = strings.TrimSpace(password)

				wallet, err := wallet.GenerateWalletFromMnemonic(ctx, loggedInUser, password, walletName)
				if err != nil {
					fmt.Println("Wallet generation failed:", err)
				} else {
					fmt.Println("Wallet generated successfully!")
					fmt.Printf("Wallet Name: %s\n", walletName)
					fmt.Printf("Mnemonic (store safely!): %s\n", string(wallet.EncryptedMnemonic))
					// You can print wallet address or keys here as needed
				}

			case "2":
				loggedInUser = nil
				fmt.Println("Logged out successfully.")

			case "3":
				fmt.Print("Enter your password to decrypt wallets: ")
				password, _ := reader.ReadString('\n')
				password = strings.TrimSpace(password)

				wallets, err := wallet.GetUserWallets(ctx, loggedInUser, password)
				if err != nil {
					fmt.Println("Failed to get wallets:", err)
				} else {
					fmt.Printf("You have %d wallets:\n", len(wallets))
					for i, w := range wallets {
						fmt.Printf("%d. %s\n", i+1, w.WalletName)
					}
				}
			case "4":
				fmt.Print("Enter wallet name to view details: ")
				walletName, _ := reader.ReadString('\n')
				walletName = strings.TrimSpace(walletName)

				fmt.Print("Enter your password: ")
				password, _ := reader.ReadString('\n')
				password = strings.TrimSpace(password)

				wallet, privKey, mnemonic, err := wallet.GetWalletDetails(ctx, loggedInUser, walletName, password)
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Wallet Details:")
					fmt.Printf("Name       : %s\n", wallet.WalletName)
					fmt.Printf("Address    : %s\n", wallet.Address)
					fmt.Printf("Public Key : %s\n", wallet.PublicKey)
					fmt.Printf("Private Key: %x\n", privKey)
					fmt.Printf("Mnemonic   : %s\n", mnemonic)
					fmt.Printf("Created At : %s\n", wallet.CreatedAt.Format(time.RFC3339))
				}

			case "5":
				fmt.Println("Exiting Wallet CLI.")
				return

			case "6":
				checkTokenBalance(ctx, loggedInUser)

			case "7":
				transferTokens(ctx, loggedInUser)

			case "8":
				stakeTokens(ctx, loggedInUser)

			default:
				fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
			}
		}
	}
}

func checkTokenBalance(ctx context.Context, user *wallet.User) {
	fmt.Println("=== Check Token Balance ===")

	// Get wallet name
	fmt.Print("Enter wallet name: ")
	walletName := readLine()

	// Get password
	fmt.Print("Enter password: ")
	password := readLine()

	// Get token symbol
	fmt.Print("Enter token symbol (e.g., BHX): ")
	tokenSymbol := readLine()

	balance, err := wallet.CheckTokenBalance(ctx, user, walletName, password, tokenSymbol)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Balance of %s: %d tokens\n", tokenSymbol, balance)
}

func transferTokens(ctx context.Context, user *wallet.User) {
	fmt.Println("=== Transfer Tokens ===")

	// Get wallet name
	fmt.Print("Enter your wallet name: ")
	walletName := readLine()

	// Get password
	fmt.Print("Enter password: ")
	password := readLine()

	// Get recipient address
	fmt.Print("Enter recipient address: ")
	toAddress := readLine()

	// Get token symbol
	fmt.Print("Enter token symbol (e.g., BHX): ")
	tokenSymbol := readLine()

	// Get amount
	fmt.Print("Enter amount to transfer: ")
	amountStr := readLine()
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid amount: %v\n", err)
		return
	}

	err = wallet.TransferTokens(ctx, user, walletName, password, toAddress, tokenSymbol, amount)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully transferred %d %s tokens to %s\n", amount, tokenSymbol, toAddress)
}

func stakeTokens(ctx context.Context, user *wallet.User) {
	fmt.Println("=== Stake Tokens ===")

	// Get wallet name
	fmt.Print("Enter your wallet name: ")
	walletName := readLine()

	// Get password
	fmt.Print("Enter password: ")
	password := readLine()

	// Get token symbol
	fmt.Print("Enter token symbol (e.g., BHX): ")
	tokenSymbol := readLine()

	// Get amount
	fmt.Print("Enter amount to stake: ")
	amountStr := readLine()
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		fmt.Printf("Invalid amount: %v\n", err)
		return
	}

	err = wallet.StakeTokens(ctx, user, walletName, password, tokenSymbol, amount)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully staked %d %s tokens\n", amount, tokenSymbol)
}
