package main

import (
	"fmt"
	"os"

	"github.com/nimbo1999/financeiro/authentication/pkg/crypto"
)

func main() {
	privateKey, publicKey, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("private_key.pem", privateKey, 0600); err != nil {
		panic(err)
	}
	if err := os.WriteFile("public_key.pem", publicKey, 0644); err != nil {
		panic(err)
	}
	fmt.Println("Keys generated and saved to files.")
}
