package main

import (
	"log"
	"os"

	"github.com/MakayaYoel/CoinTracker/bot"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Please create an environment file (.env) to use this discord bot.")
	}
}

func main() {
	// Load environment variables
	botToken, ok := os.LookupEnv("DISCORD_BOT_TOKEN")
	if !ok {
		log.Fatal("Please specify a DISCORD_BOT_TOKEN in your environment file.")
	}

	cgToken, ok := os.LookupEnv("COINGECKO_API_KEY")
	if !ok {
		log.Fatal("Please specify a COINGECKO_API_KEY in your environment file.")
	}

	guildID, ok := os.LookupEnv("DISCORD_GUILD_ID")
	if !ok {
		log.Fatal("Please specify a DISCORD_GUILD_ID in your environment file.")
	}

	// Start discord bot.
	bot.BotToken = botToken
	bot.CGToken = cgToken
	bot.GuildID = guildID

	bot.Start()
}
