package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/joho/godotenv"
	cg "github.com/superoo7/go-gecko/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var tokenList = make(map[string]string)

type Token struct {
	Name    string   `json:"name"`
	Tickers []Ticker `json:"tickers"`
}

type Ticker struct {
	Base                   string             `json:"base"`
	Target                 string             `json:"target"`
	Market                 Market             `json:"market"`
	Last                   float64            `json:"last"`
	Volume                 float64            `json:"volume"`
	ConvertedLast          map[string]float64 `json:"converted_last"`
	ConvertedVolume        map[string]float64 `json:"converted_volume"`
	TrustScore             string             `json:"trust_score"`
	BidAskSpreadPercentage float64            `json:"bid_ask_spread_percentage"`
	Timestamp              string             `json:"timestamp"`
	LastTradedAt           string             `json:"last_traded_at"`
	LastFetchAt            string             `json:"last_fetch_at"`
	IsAnomaly              bool               `json:"is_anomaly"`
	IsStale                bool               `json:"is_stale"`
	TradeURL               string             `json:"trade_url"`
	TokenInfoURL           string             `json:"token_info_url"`
	CoinID                 string             `json:"coin_id"`
	TargetCoinID           string             `json:"target_coin_id"`
}

type Market struct {
	Name                string `json:"name"`
	Identifier          string `json:"identifier"`
	HasTradingIncentive bool   `json:"has_trading_incentive"`
}

func init() {
	var err error
	err = godotenv.Load("dev.env")
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	log := zap.NewProductionEncoderConfig()
	log.EncodeLevel = zapcore.CapitalLevelEncoder
	log.EncodeTime = zapcore.RFC3339TimeEncoder

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(log), os.Stdout, zap.InfoLevel))

	updater, err := gotgbot.NewUpdater(logger, os.Getenv("TG_KEY"))
	if err != nil {
		logger.Panic("Updater failed to start")
		return
	}

	loadTokenList()
	logger.Sugar().Info("Tokens list loaded")

	logger.Sugar().Info("Updater started successfully")
	updater.StartCleanPolling()
	//updater.Dispatcher.AddHandler(handlers.NewCommand("romestime", romesTime))
	//updater.Dispatcher.AddHandler(handlers.NewCommand("price", usdPrice))

	// reply to messages satisfying this regex
	updater.Dispatcher.AddHandler(handlers.NewRegex("(?i)/", returnTokenPrice))
	updater.Dispatcher.AddHandler(handlers.NewRegex("(?i)arb", executArbitrage))
	updater.Idle()
}

func loadTokenList() {
	tokenList["xor"] = "sora"
	tokenList["val"] = "sora-validator-token"
	tokenList["link"] = "chainlink"
	tokenList["ramp"] = "ramp"
	tokenList["shitcoin"] = "shitcoin"
}

func getTokenId(tokenTicker string) string {
	if _, ok := tokenList[tokenTicker]; ok {
		return tokenList[tokenTicker]
	}
	return tokenList["shitcoin"]
}

func getTokenPrice(tokenName string) string {
	cg := cg.NewClient(nil)
	price, err := cg.SimpleSinglePrice(tokenList[tokenName[1:]], "usd")
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%s is worth %f %s", tokenName[1:], price.MarketPrice, price.Currency)
}

func getTokenArbitrage(tokenName string) string {

	endpointURL := "https://api.coingecko.com/api/v3/coins/" + tokenList[tokenName[3:]] + "/tickers"
	resp, err := http.Get(endpointURL)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var token Token
	json.Unmarshal(bodyBytes, &token)
	//fmt.Printf("API Response as struct %+v\n", token)

	return buildArbitrageMessage(token)
}

func buildArbitrageMessage(token Token) string {
	var message string = "\n"
	counter := 1
	for _, ticker := range token.Tickers {
		//for _, market := range ticker
		message += " " + ticker.Market.Name + "\n"
		for _, value := range ticker.ConvertedLast {
			if counter == 1 {
				message += "" + fmt.Sprintf("BTC: %f", value) + "\n"
			}
			if counter == 2 {
				message += "" + fmt.Sprintf("ETH: %f", value) + "\n"
			}
			if counter == 3 {
				message += "" + fmt.Sprintf("USD: %f", value) + "\n"
			}
			counter++
		}
		counter = 1
	}
	return message
}

func returnTokenPrice(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, getTokenPrice(u.EffectiveMessage.Text))
	return nil
}

func executArbitrage(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, getTokenArbitrage(u.EffectiveMessage.Text))
	return gotgbot.ContinueGroups{} // will keep executing handlers, even after having been caught by this one.
}
