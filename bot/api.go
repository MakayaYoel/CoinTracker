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

// maxCurrenciesPerPage represents the max currencies displayed in the "currencies" command pagination widget.
const maxCurrenciesPerPage = 9

// GetCoinPrice returns a discordgo.MessageEmbed containing the specified coin's price in the also specified currency.
func GetCoinPrice(id string, currency string) ([]*discordgo.MessageEmbed, error) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s&include_market_cap=false&include_24hr_vol=false&include_24hr_change=false&include_last_updated_at=true&precision=full", id, currency)
	responseData := getResponseData(url, map[string]string{"x-cg-demo-api-key": CGToken})

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

// GetCurrencies returns a slice of MessageEmbeds containing the available currencies.
func GetCurrencies() ([]*discordgo.MessageEmbed, error) {
	responseData := getResponseData("https://api.coingecko.com/api/v3/simple/supported_vs_currencies", nil)
	data, err := cnvToEmbedSlice(responseData.([]interface{}))

	if err != nil {
		return nil, fmt.Errorf("could not retrieve currencies: %s", err.Error())
	}

	var pages []*discordgo.MessageEmbed

	for len(data) != 0 {
		maxAmount := int(math.Min(maxCurrenciesPerPage, float64(len(data))))

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

	return pages, nil
}

// cnvToEmbedSlice converts the specified interface slice into a MessageEmbedField slice.
func cnvToEmbedSlice(data []interface{}) ([]*discordgo.MessageEmbedField, error) {
	result := make([]*discordgo.MessageEmbedField, len(data))

	for i, v := range data {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("could not convert data at index %d to embed slice", i)
		}

		result[i] = &discordgo.MessageEmbedField{
			Name:   strings.ToUpper(str),
			Inline: true,
		}
	}

	return result, nil
}

// getResponseData() returns an interface representing the response data of an HTTP GET request sent to the specified url with the also specified headers.
func getResponseData(url string, headers map[string]string) interface{} {
	request, _ := http.NewRequest("GET", url, nil)

	request.Header.Add("accept", "application/json")

	for h, hd := range headers {
		request.Header.Add(h, hd)
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Fatalf("could not complete request: %s", err.Error())
	}

	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var responseData interface{}
	json.Unmarshal([]byte(body), &responseData)

	return responseData
}
