package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var bot *Bot

func init() {
	bot = NewBot(config.Email, config.Password, config.Token, config.CommandPrefix)
	bot.RegisterCommand("eightball", cmdEightball)
}

func main() {
	err := bot.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	log.Println("ready")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
