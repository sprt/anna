package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/spf13/viper"

	"github.com/sprt/anna"
	"github.com/sprt/anna/commands"
	"github.com/sprt/anna/tasks"
)

var (
	dbPath     string
	configPath string
)

func init() {
	flag.StringVar(&dbPath, "db", "", "path to the database")
	flag.StringVar(&configPath, "conf", "", "path to the ini config file")
}

func registerCommands(bot *anna.Bot) {
	bot.RegisterCommand("eightball", commands.Eightball)
	bot.RegisterCommand("friends", commands.Friends)
	bot.RegisterCommand("ping", commands.Ping)
	bot.RegisterCommand("players", commands.Players)
	bot.RegisterCommand("random", commands.Random)
	bot.RegisterCommand("userinfo", commands.UserInfo)
}

func registerTasks(bot *anna.Bot) {
	bot.RegisterTask(tasks.AcceptFriendRequests, 3*time.Hour)
	bot.RegisterTask(tasks.FetchMembers, 6*time.Hour)
	bot.RegisterTask(tasks.FetchOnlineFriends, time.Minute)
	// bot.RegisterTask(tasks.SendFriendRequests, 3*time.Hour)
}

func main() {
	flag.Parse()

	switch {
	case dbPath == "", configPath == "":
		flag.Usage()
		os.Exit(2)
	}

	configPath, err := filepath.Abs(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	configDir, configFile := path.Split(configPath)
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(strings.TrimSuffix(configFile, path.Ext(configFile)))
	v.AddConfigPath(configDir)
	if err := v.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config := &anna.Config{
		CommandPrefix:    v.GetString("command_prefix"),
		FriendRequestMsg: v.GetString("friend_request_msg"),

		DiscordEmail:    v.GetString("discord.email"),
		DiscordPassword: v.GetString("discord.password"),
		DiscordToken:    v.GetString("discord.token"),
		OwnerID:         v.GetString("discord.owner_id"),

		SocialClubCrewID: v.GetInt("socialclub.crew_id"),

		GoogleDriveRosterID: v.GetString("roster.sheet_id"),

		PSNEmail:        v.GetString("psn.email"),
		PSNUsername:     v.GetString("psn.username"),
		PSNPassword:     v.GetString("psn.password"),
		PSNClientID:     v.GetString("psn.client_id"),
		PSNClientSecret: v.GetString("psn.client_secret"),
		PSNDuid:         v.GetString("psn.duid"),
	}

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	bot := anna.NewBot(config, db)
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
