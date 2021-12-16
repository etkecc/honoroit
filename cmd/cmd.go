package main

import (
	"fmt"

	"gitlab.com/etke.cc/honoroit/matrix"
)

func main() {
	bot, err := matrix.NewBot("", "", "", "")
	if err != nil {
		panic(err)
	}
	fmt.Println("bot has been created")
	if err = bot.WithStore(); err != nil {
		panic(err)
	}
	fmt.Println("account data initialized")

	if err = bot.Run(); err != nil {
		panic(err)
	}
}
