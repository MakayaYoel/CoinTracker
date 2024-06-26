package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
)

var (
	GuildID  string
	BotToken string
	CGToken  string

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Shows you all of CoinTracker's commands.",
		},
		{
			Name:        "getcoinprice",
			Description: "Returns the current price of a specified coin.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The coin's id (e.g. Bitcoin, Ethereum)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currency",
					Description: "The currency you want the information on (e.g. CAD, USD)",
					Required:    true,
				},
			},
		},
		{
			Name:        "currencies",
			Description: "Returns a list of valid currencies.",
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
									Value:  "Shows you all of CoinTracker's commands.",
									Inline: false,
								},
								{
									Name:   "/getcoinprice <id> <currency>",
									Value:  "Returns the current price of a specified coin.",
									Inline: false,
								},
								{
									Name:   "/currencies",
									Value:  "Returns a list of valid currencies.",
									Inline: false,
								},
								{
									Name:   "Credits",
									Value:  "This bot is powered by the CoinGecko API.",
									Inline: false,
								},
							},
						},
					},
				},
			})
		},

		"getcoinprice": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options

			var content string
			embeds, err := GetCoinPrice(options[0].StringValue(), options[1].StringValue())

			if err != nil {
				content = fmt.Sprintf("There was an error trying to complete the request: %s.", err.Error())
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
					Embeds:  embeds,
				},
			})
		},

		"currencies": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Creating pagination widget...",
				},
			})

			p := dgwidgets.NewPaginator(s, i.ChannelID)

			p.Add(GetCurrencies()...)
			p.SetPageFooters()

			p.Widget.Timeout = time.Second * 45
			p.Widget.UserWhitelist = []string{i.Member.User.ID}

			p.Spawn()
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

	GetCurrencies()

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
