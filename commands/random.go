package commands

import (
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
)

func Random(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	if len(args) < 2 {
		return nil
	}
	_, err := bot.Session.ChannelMessageSend(msg.ChannelID, args[rand.Intn(len(args))])
	return err
}
