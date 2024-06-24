package bot

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildID  string
	BotToken string
	CGToken  string

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Learn what are my commands!",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Type:        discordgo.EmbedTypeRich,
							Title:       "CoinTracker - Commands",
							Description: "Here are my commands!",
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:   "/help",
									Value:  "Shows you all my commands!",
									Inline: false,
								},
							},
						},
					},
				},
			})
		},
	}
)

func Start() {
	s, err := discordgo.New("Bot " + BotToken)

	if err != nil {
		log.Fatalf("There was an error creating a session for the bot: %s", err.Error())
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Redirect to the command's handler
		if cmdHandler, exists := commandHandlers[i.ApplicationCommandData().Name]; exists {
			cmdHandler(s, i)
		}
	})

	// Start bot
	err = s.Open()

	if err != nil {
		log.Fatalf("There was an error trying to start the bot: %s", err.Error())
	}

	log.Printf("Started bot on %s#%s...", s.State.User.Username, s.State.User.Discriminator)

	defer s.Close()

	// Register commands
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)

		if err != nil {
			log.Printf("Failed to register the '%s' command: %s", v.Name, err.Error())
		} else {
			registeredCommands[i] = cmd
		}
	}

	// Wait for interrupt signal to turn off the bot
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Print("Turning off bot...")

	// Unregister commands
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)

		if err != nil {
			log.Panicf("There was an error trying to delete the '%s' command: %s", v.Name, err.Error())
		}
	}
}
