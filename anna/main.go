package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"

	"github.com/sprt/anna"
	"github.com/sprt/anna/commands"
	"github.com/sprt/anna/tasks"
)

var (
	dbPath string
)

func init() {
	flag.StringVar(&dbPath, "db", "", "path to the database")
}

func registerCommands(bot *anna.Bot) {
	bot.RegisterCommand("eightball", commands.Eightball)
	bot.RegisterCommand("userinfo", commands.UserInfo)
}

func registerTasks(bot *anna.Bot) {
	bot.RegisterTask(tasks.FetchMembers, time.Hour)
}

func main() {
	flag.Parse()

	if dbPath == "" {
		flag.Usage()
		os.Exit(2)
	}
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	bot := anna.NewBot(db)
	registerCommands(bot)
	registerTasks(bot)

	err = bot.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	log.Print("ready")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
