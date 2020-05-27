package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ctx context.Context = context.Background()
var opt, err = redis.ParseURL(os.Getenv("REDIS_URL"))
var client *redis.Client = redis.NewClient(&redis.Options{
	Addr:     opt.Addr,
	Password: opt.Password,
	DB:       opt.DB,
})

func main() {

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder

	// app, err := newrelic.NewApplication(
	// 	newrelic.Config{AppName: "mighty-gta-bot-v2", License: os.Getenv("NEW_RELIC_LICENSE_KEY"), Logger: newrelic.NewLogger(os.Stdout) },

	// )

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), os.Stdout, zap.InfoLevel))
	defer logger.Sync() // flushes buffer, if any
	l := logger.Sugar()

	l.Info("Starting gotgbot...")
	fmt.Println(opt.Password)

	if err != nil {
		l.Fatalw("failed to start updater", zap.Error(err))
	}

	token := os.Getenv("TOKEN")
	port, err := strconv.Atoi(os.Getenv("PORT"))

	fmt.Println(port)

	if port == 0 {
		port = 8080
	}

	// port:=5000
	fmt.Println(port)

	updater, err := gotgbot.NewUpdater(token, logger)
	if err != nil {
		l.Fatalw("failed to start updater", zap.Error(err))
	}

	// /start Handler
	updater.Dispatcher.AddHandler(handlers.NewCommand("start", startHandler))

	// /status Handler
	updater.Dispatcher.AddHandler(handlers.NewCommand("status", statusHandler))

	// /subscription Handler
	updater.Dispatcher.AddHandler(handlers.NewCommand("subscribe", subscriptionHandler))
	updater.Dispatcher.AddHandler(handlers.NewCommand("unsubscribe", unsubscribeHandler))

	if os.Getenv("USE_WEBHOOKS") == "t" {
		fmt.Println("Using Webhooks")
		// start getting updates
		webhook := gotgbot.Webhook{
			Serve:          "0.0.0.0",
			ServePort:      port,
			ServePath:      updater.Bot.Token,
			URL:            os.Getenv("WEBHOOK_URL"),
			MaxConnections: 30,
		}
		updater.StartWebhook(webhook)
		ok, err := updater.SetWebhook(updater.Bot.Token, webhook)
		if err != nil {
			l.Fatalw("Failed to start bot", zap.Error(err))
		}
		if !ok {
			l.Fatalw("Failed to set webhook", zap.Error(err))
		}
	} else {
		err := updater.StartPolling()
		if err != nil {
			l.Fatalw("Failed to start polling", zap.Error(err))
		}
	}

	go SubscriptionStatusChecker(ctx, *updater.Dispatcher.Bot, client)
	// wait
	updater.Idle()

}

func startHandler(b ext.Bot, u *gotgbot.Update) error {
	b.SendMessage(u.Message.Chat.Id, "Welcome to the MightyGTA Bot!")
	return nil
}

func statusHandler(b ext.Bot, u *gotgbot.Update) error {

	rockstar := checkStatus()

	statusMessage := fmt.Sprintf("%s servers are %s with status-code %d", rockstar.Name, rockstar.StatusTag, rockstar.StatusCode)

	_, err := b.ReplyText(u.EffectiveChat.Id, statusMessage, u.EffectiveMessage.MessageId)
	if err != nil {
		b.Logger.Warnw("Error sending V2", zap.Error(err))
	}

	return nil
}

func subscriptionHandler(b ext.Bot, u *gotgbot.Update) error {

	var msg string

	chatID := u.Message.Chat.Id
	stringChatID := strconv.Itoa(chatID)

	checkKey, err := client.Exists(ctx, stringChatID).Result()

	if checkKey == int64(1) {
		msg = "You are already subscribed to my services. If you want to unsubscribe, use /unsubscribe."

	} else {
		_, err := client.Set(ctx, stringChatID, chatID, 0).Result()

		if err != nil {
			b.Logger.Warnw("Error Setting Key", zap.Error(err))
		}
		msg = "Thank you for the subscription to the bot. I will notify you whenever there is any change on GTA online servers."
		fmt.Println(chatID, "has been added to db")
	}
	_, err = b.SendMessage(chatID, msg)

	if err != nil {
		b.Logger.Warnw("Error sending V2", zap.Error(err))
	}
	return nil
}

func unsubscribeHandler(b ext.Bot, u *gotgbot.Update) error {
	var msg string

	chatID := u.Message.Chat.Id
	stringChatID := strconv.Itoa(chatID)

	checkKey, err := client.Exists(ctx, stringChatID).Result()

	if checkKey == int64(1) {
		msg = "You have been unsubscribed. To again subscribe, use /subscribe"
		_, err := client.Del(ctx, stringChatID).Result()
		fmt.Println(chatID, "has been deleted")

		if err != nil {
			b.Logger.Warnw("Error Deleting Entry", zap.Error(err))
		}

	} else {
		msg = "You are not subscribed with me. I cannot unsubscribe you."
	}
	_, err = b.SendMessage(chatID, msg)

	if err != nil {
		b.Logger.Warnw("Error sending V2", zap.Error(err))
	}
	return nil
}
