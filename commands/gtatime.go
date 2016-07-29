package commands

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sprt/anna"
)

func GTATime(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	t := time.Unix(75600+(time.Now().UTC().Unix()-1452351963)*30, 0).UTC()
	_, err := bot.Session.ChannelMessageSend(msg.ChannelID, t.Format("15:04"))
	return err
}
