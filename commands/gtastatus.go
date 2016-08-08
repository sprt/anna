package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/sprt/anna"
)

func GTAStatus(bot *anna.Bot, msg *discordgo.Message, args []string) error {
	err := bot.Session.ChannelTyping(msg.ChannelID)
	if err != nil {
		return err
	}
	psn := bot.PSN.Status()
	rockstar := bot.SocialClub.Status()
	say := fmt.Sprintf("PSN: %s, R* services: %s", string(psn), string(rockstar))
	_, err = bot.Session.ChannelMessageSend(msg.ChannelID, say)
	return err
}
