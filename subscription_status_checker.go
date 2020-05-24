package main

import (
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// SubscriptionStatusChecker checks the status of the Rockstar and sends the message to the stored ChatID
func SubscriptionStatusChecker(ctx context.Context, b ext.Bot, client *redis.Client) {

	statusCode := 1
	fmt.Println("Started Subscription Checker")

	for {

		rockstar := checkStatus()
		fmt.Println("Checking Again...")

		if statusCode != rockstar.StatusCode {

			keys, err := client.Keys(ctx, "*").Result()

			if err != nil {
				b.Logger.Warnw("Error Fetching All Keys", zap.Error(err))
			}

			// Iterate through the Stored ChatIDs and Send messages
			for _, key := range keys {

				stringChatID, err := client.Get(ctx, key).Result()
				if err != nil {
					b.Logger.Warnw("Error Getting the value", zap.Error(err))
				}

				chatID, err := strconv.Atoi(stringChatID)
				if err != nil {
					b.Logger.Warnw("Error converting ChatID from ChatIDByte", zap.Error(err))
				}

				go b.SendMessage(chatID, "GTA Online Servers are now "+rockstar.StatusTag)
			}

			// Update the Status Code
			statusCode = rockstar.StatusCode

		} else {

			fmt.Println("Status Checked. There is no change.")

		}

		time.Sleep(time.Minute * 10)
	}

}
