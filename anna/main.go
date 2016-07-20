package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sprt/anna"
	"github.com/sprt/anna/commands"
)

var bot *anna.Bot

func init() {
	bot = anna.NewBot()
	bot.RegisterCommand("eightball", commands.Eightball)
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
