package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func GetCoinPrice(id string, currency string) ([]*discordgo.MessageEmbed, error) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s&include_market_cap=false&include_24hr_vol=false&include_24hr_change=false&include_last_updated_at=true&precision=full", id, currency)

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("accept", "application/json")
	request.Header.Add("x-cg-demo-api-key", CGToken)

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, fmt.Errorf("could not complete request")
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var responseData interface{}
	json.Unmarshal([]byte(body), &responseData)

	// type assert
	rData, ok := responseData.(map[string]interface{})

	if !ok || len(rData) == 0 {
		return nil, fmt.Errorf("coin was not found")
	}

	var coinName, currentPrice, lastUpdatedAt string

	for c, d := range rData {
		cD := d.(map[string]interface{})

		currency, ok := cD[currency]

		if !ok {
			return nil, fmt.Errorf("could not find currency")
		}

		caser := cases.Title(language.AmericanEnglish)
		p := message.NewPrinter(language.AmericanEnglish)

		coinName = caser.String(c)
		currentPrice = p.Sprintf("%.2f$", currency.(float64))
		lastUpdatedAt = time.Unix(int64(cD["last_updated_at"].(float64)), 0).Format("Mon, 01/02/2006, 03:04:05 PM")
	}

	return []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       fmt.Sprintf("Current Coin Price - %s", coinName),
			Description: fmt.Sprintf("Currently showing the coin price for %s.", coinName),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Currency",
					Value:  strings.ToUpper(currency),
					Inline: true,
				},
				{
					Name:   "Current Price",
					Value:  currentPrice,
					Inline: true,
				},
				{
					Name:   "Last Updated At",
					Value:  lastUpdatedAt,
					Inline: false,
				},
			},
		},
	}, nil
}

func GetCurrencies() []*discordgo.MessageEmbed {
	url := "https://api.coingecko.com/api/v3/simple/supported_vs_currencies"

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("accept", "application/json")

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Fatal("for later 1")
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var responseData interface{}
	json.Unmarshal([]byte(body), &responseData)

	data, err := cnvToEmbedSlice(responseData.([]interface{}))

	if err != nil {
		log.Fatal("for later 2")
	}

	var pages []*discordgo.MessageEmbed

	for len(data) != 0 {
		maxAmount := int(math.Min(9, float64(len(data))))

		// Get page, update data
		embed := &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Currencies",
			Description: "Here are all the available currencies!",
			Fields:      data[:maxAmount],
		}

		pages = append(pages, embed)
		data = data[maxAmount:]
	}

	return pages
}

func cnvToEmbedSlice(data []interface{}) ([]*discordgo.MessageEmbedField, error) {
	result := make([]*discordgo.MessageEmbedField, len(data))

	for i, v := range data {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("could not convert element at index %d to string", i)
		}

		result[i] = &discordgo.MessageEmbedField{
			Name:   strings.ToUpper(str),
			Inline: true,
		}
	}

	return result, nil
}
