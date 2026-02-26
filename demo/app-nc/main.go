package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func main() {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 31

	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		result[i] = charset[n.Int64()]
	}

	// Read an input string from the user
	fmt.Printf("Benvenuto nel servizio di Vincenzino per prima cosa fai questo")
	fmt.Println("Enter a string:")
	var input string
	fmt.Scanln(&input)
	fmt.Println("You entered:", input)

	final := string(result) + "="
	fmt.Println("Take a cookie man: " + final)
}
