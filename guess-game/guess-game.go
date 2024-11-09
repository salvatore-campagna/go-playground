// This program is a simple number guessing game where the computer tries to guess a
// target number provided as a command-line argument. The target number must be between 1 and 100.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing number. Please provide a valid integer")
		return
	}

	var trials, lower, upper = 1, 1, 100
	target, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid number. Please provide a valid integer")
		return
	}

	if target < 1 || target > 100 {
		fmt.Printf("Invalid number. Must be in range [ %v, %v ]\n", lower, upper)
		return
	}

	for guess := rand.Intn(upper-lower+1) + lower; guess != target; trials++ {
		fmt.Printf("guess (%v) => %v\n", trials, guess)
		if guess > target {
			upper = guess - 1
		} else if guess < target {
			lower = guess + 1
		}
		guess = rand.Intn(upper-lower+1) + lower
	}

	fmt.Printf("Took me %v trials to guess %v\n", trials, target)
}
