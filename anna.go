package anna

import (
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	email, password, token string
	cmdPrefix              string
	db                     *bolt.DB
	Session                *discordgo.Session
	user, owner            *discordgo.User
	commands               []*command
	tasks                  []*task
}

func NewBot(db *bolt.DB) *Bot {
	return &Bot{
		email:     Config.DiscordEmail,
		password:  Config.DiscordPassword,
		token:     Config.DiscordToken,
		cmdPrefix: Config.CommandPrefix,
		db:        db,
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

	b.Session.AddHandler(b.onReady)
	b.Session.AddHandler(b.onMessageCreate)

	return nil
}

func (b *Bot) RegisterCommand(name string, fn func(*Bot, *discordgo.Message, []string)) {
	// FIXME: panic if a task with the same name already exists
	b.commands = append(b.commands, &command{
		name: name,
		fn:   fn,
	})
}

func (b *Bot) RegisterTask(fn func(*Bot), sleep time.Duration) {
	b.tasks = append(b.tasks, &task{
		fn:    fn,
		sleep: sleep,
	})
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	for _, task := range b.tasks {
		go func() {
			for {
				task.fn(b)
				time.Sleep(task.sleep)
			}
		}()
	}
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
		if strings.EqualFold(cmd.name, name) {
			cmd.fn(b, message, args)
			break
		}
	}
}

type command struct {
	name string
	fn   func(*Bot, *discordgo.Message, []string)
}

type task struct {
	fn    func(*Bot)
	sleep time.Duration
}
