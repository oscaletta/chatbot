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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Time struct {
	Abbreviation string      `json:"abbreviation"`
	ClientIP     string      `json:"client_ip"`
	Datetime     string      `json:"datetime"`
	DayOfWeek    int64       `json:"day_of_week"`
	DayOfYear    int64       `json:"day_of_year"`
	Dst          bool        `json:"dst"`
	DstFrom      interface{} `json:"dst_from"`
	DstOffset    int64       `json:"dst_offset"`
	DstUntil     interface{} `json:"dst_until"`
	RawOffset    int64       `json:"raw_offset"`
	Timezone     string      `json:"timezone"`
	Unixtime     int64       `json:"unixtime"`
	UTCDatetime  string      `json:"utc_datetime"`
	UTCOffset    string      `json:"utc_offset"`
	WeekNumber   int64       `json:"week_number"`
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

	logger.Sugar().Info("Updater started successfully")
	updater.StartCleanPolling()
	updater.Dispatcher.AddHandler(handlers.NewCommand("start", start))
	updater.Dispatcher.AddHandler(handlers.NewCommand("rome", rome))
	updater.Dispatcher.AddHandler(handlers.NewCommand("dumb", hthau))

	//updater.Dispatcher.AddHandler(handlers.NewMessage(Filters.Text, echo))

	// reply to messages satisfying this regex
	updater.Dispatcher.AddHandler(handlers.NewRegex("(?i)llamas", replyMessage))
	updater.Idle()

}

func price(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, "The current price is 10")
	return nil
}

func echo(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.EffectiveChat.Id, u.EffectiveMessage.Text)
	return nil
}

func start(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, "Congrats! You just issued a /start on your go bot!")
	return nil
}

func rome(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, getTime())
	return nil
}

func hthau(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, "Who the hell are you?")
	return nil
}

func replyMessage(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, "Me llamo Lucia Romano")
	return gotgbot.ContinueGroups{} // will keep executing handlers, even after having been caught by this one.
}

func getTime() string {
	fmt.Println("1. Performing Http Get...")
	resp, err := http.Get("http://worldtimeapi.org/api/timezone/Europe/Rome")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to Todo struct
	var timeStruct Time
	json.Unmarshal(bodyBytes, &timeStruct)
	fmt.Printf("API Response as struct %+v\n", timeStruct)
	return timeStruct.Datetime
}
