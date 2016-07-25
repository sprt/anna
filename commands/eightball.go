package commands

import (
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	"github.com/sprt/anna"
)

var eightballAnswers = []string{
	"It is certain",
	"It is decidedly so",
	"Without a doubt",
	"Yes, definitely",
	"You may rely on it",
	"As I see it, yes",
	"Most likely",
	"Outlook good",
	"Yes",
	"Signs point to yes",
	"Reply hazy, try again",
	"Ask again later",
	"Better not tell you now",
	"Cannot predict now",
	"Concentrate and ask again",
	"Don't count on it",
	"My reply is no",
	"My sources say no",
	"Outlook not so good",
	"Very doubtful",
}

func Eightball(bot *anna.Bot, msg *discordgo.Message, args []string) {
	answer := eightballAnswers[rand.Intn(len(eightballAnswers))]
	if _, err := bot.Session.ChannelMessageSend(msg.ChannelID, answer); err != nil {
		log.Printf("eightball: %s", err)
	}
}
