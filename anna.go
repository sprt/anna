package anna

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type command struct {
	name string
	fn   func(*Bot, *discordgo.Message, []string)
}

type Bot struct {
	email, password, token string
	cmdPrefix              string
	Session                *discordgo.Session
	user, owner            *discordgo.User
	commands               []*command
}

func NewBot() *Bot {
	return &Bot{
		email:     Config.Email,
		password:  Config.Password,
		token:     Config.Token,
		cmdPrefix: Config.CommandPrefix,
	}
}

func (b *Bot) Start() error {
	session, err := discordgo.New(b.email, b.password, b.token)
	if err != nil {
		return err
	}

	user, err := session.User("@me")
	if err != nil {
		return err
	}

	err = session.Open()
	if err != nil {
		return err
	}

	b.Session = session
	b.user = user
	// TODO: initialize owner

	b.Session.AddHandler(b.onMessageCreate)

	return nil
}

func (b *Bot) RegisterCommand(name string, fn func(*Bot, *discordgo.Message, []string)) {
	b.commands = append(b.commands, &command{
		name: name,
		fn:   fn,
	})
}

func (b *Bot) onMessageCreate(s *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == b.user.ID {
		return
	}

	if strings.HasPrefix(mc.Content, b.cmdPrefix) {
		args := strings.Fields(mc.Content[len(b.cmdPrefix):])
		b.onCommand(mc.Message, args[0], args[1:])
		return
	}
}

func (b *Bot) onCommand(message *discordgo.Message, name string, args []string) {
	for _, cmd := range b.commands {
		if cmd.name == strings.ToLower(name) {
			cmd.fn(b, message, args)
			break
		}
	}
}
