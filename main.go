package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/justinian/dice"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Token string
}

var helpMsg = strings.Join([]string{
	"Hi! I'm a dice bot.",
	"Use `!roll <format>` to roll some dice.",
	"For format help, see: http://bit.ly/dice-bot",
}, "\n")

func main() {
	var c Config
	envconfig.Process("discordice", &c)

	if c.Token == "" {
		log.Fatal("No token provided.")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		log.Fatalf("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %s", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Print("Discordice is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Print("Discord session ready.")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!rollhelp") {
		c, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			log.Printf("Couldn't open DM to user: %v", err)
			return
		}

		_, err = s.ChannelMessageSend(c.ID, helpMsg)
		if err != nil {
			log.Printf("Error sending message to %v: %s", c, err)
			return
		}

		mc, err := s.State.Channel(m.ChannelID)
		if err == nil && mc.Type == discordgo.ChannelTypeGuildText {
			err = s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				log.Printf("Error deleting message in %v: %s", mc, err)
				return
			}
		}
	} else if strings.HasPrefix(m.Content, "!roll") {
		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			log.Printf("No channel ID: %v", err)
			return
		}

		result, reason, err := dice.Roll(m.Content)
		if err != nil {
			log.Printf("Dice error: %s", err)
			return
		}

		if reason != "" {
			reason = fmt.Sprintf(" *%s*", reason)
		}

		resultStrs := strings.Split(result.String(), "\n")
		for i, s := range resultStrs {
			if i == 0 {
				resultStrs[i] = fmt.Sprintf("*%s*", strings.TrimSpace(s))
			} else {
				resultStrs[i] = fmt.Sprintf("_%s_", strings.TrimSpace(s))
			}
		}

		text := fmt.Sprintf("*<@%s>* rolled `%s`:%s\n%s",
			m.Author.ID,
			result.Description(),
			reason,
			strings.Join(resultStrs, "\n"))

		_, err = s.ChannelMessageSend(m.ChannelID, text)
		if err != nil {
			log.Printf("Error sending message to %v: %s", c, err)
			return
		}

		if c.Type == discordgo.ChannelTypeGuildText {
			err = s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				log.Printf("Error deleting message in %v: %s", c, err)
				return
			}
		}
	}
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	log.Printf("Joined Guild: %s (%s) in %s",
		event.Guild.Name, event.Guild.ID, event.Guild.Region)

	owner, err := s.User(event.Guild.OwnerID)
	if err == nil {
		log.Printf("\tOwned by: %s#%s (%s)",
			owner.Username, owner.Discriminator, owner.ID)
	}
}
