package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
)

func Ping(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", string(msg.Timestamp))
	if err != nil {
		return err
	}
	say := fmt.Sprintf("Pong! (%.3fs)", time.Now().Sub(t).Seconds())
	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, say)
	return err
}
