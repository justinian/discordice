package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/justinian/dice"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GuildID        string // Test guild ID
	Token          string // Bot API token
	RemoveCommands bool   // Remove commands on shutdown
}

func rollHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options
	optionStrs := make([]string, len(options))
	for i := range options {
		optionStrs[i] = options[i].StringValue()
	}

	result, reason, err := dice.Roll(strings.Join(optionStrs, " "))
	if err != nil {
		errMsg := fmt.Sprintf(":x: *Dice error:* %s", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errMsg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
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

	text := fmt.Sprintf("Rolling `%s`:%s\n%s",
		result.Description(),
		reason,
		strings.Join(resultStrs, "\n"))

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// Ignore type for now, they will be discussed in "responses"
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: text},
	})
}

func main() {
	var c Config
	envconfig.Process("discordice", &c)

	s, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	var rollCommand = discordgo.ApplicationCommand{
		Name:        "roll",
		Description: "Roll some dice!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "expression",
				Description: "Dice expression to roll with optional description (e.g. '2d6 damage')",
				Required:    true,
			},
		},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commandName := i.ApplicationCommandData().Name
		if commandName == "roll" {
			rollHandler(s, i)
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding command...")
	registeredCmd, err := s.ApplicationCommandCreate(s.State.User.ID, c.GuildID, &rollCommand)
	if err != nil {
		log.Panicf("Cannot create roll command: %v", err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if c.RemoveCommands {
		log.Println("Removing command...")
		err := s.ApplicationCommandDelete(s.State.User.ID, c.GuildID, registeredCmd.ID)
		if err != nil {
			log.Panicf("Cannot delete roll command: %v", err)
		}
	}

	log.Println("Gracefully shutting down.")
}
