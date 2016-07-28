package commands

import (
	"strings"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
	"github.com/sprt/anna/services/psn"
)

func Friends(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	if msg.Author.ID != bot.OwnerID || len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "add":
		return friendsAdd(bot, msg, args[1:])
	default:
		return nil
	}
}

func friendsAdd(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	if len(args) == 0 {
		return nil
	}

	err := bot.Session.ChannelTyping(msg.ChannelID)
	if err != nil {
		return err
	}

	username := args[0]
	reqMsg := strings.Join(args[1:], "")
	var say string

	if reqMsg == "" {
		reqMsg = bot.FriendRequestMsg
	}
	err = bot.PSN.SendFriendRequest(username, reqMsg)
	switch {
	case err == psn.ErrAccessDeniedPrivacy:
		say = "Denied by privacy level"
	case err == psn.ErrAlreadyFriends:
		say = "Already friends"
	case err == psn.ErrAlreadyFriendRequested:
		say = "Already sent a request"
	case err == psn.ErrUserNotFound:
		say = "Not found"
	case err != nil:
		return err
	case err == nil:
		say = "OK"
		if err := bot.DB.Update(func(tx *bolt.Tx) error {
			buk := tx.Bucket([]byte("friended"))
			return buk.Put([]byte(username), []byte(""))
		}); err != nil {
			return err
		}
	}

	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, say)
	return err
}
