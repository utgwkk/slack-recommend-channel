package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

func postMessage(api *slack.Client, channelID, text string) (string, string, error) {
	params := slack.NewPostMessageParameters()
	params.LinkNames = 1
	msgOption := slack.MsgOptionCompose(
		slack.MsgOptionUsername("今日のおすすめチャンネル"),
		slack.MsgOptionIconEmoji("tada"),
		slack.MsgOptionText(text, false),
		slack.MsgOptionPostMessageParameters(params),
	)
	return api.PostMessage(channelID, msgOption)
}

func chooseChannel(channelNames []string) string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(channelNames))
	return channelNames[index]
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func filterChannels(channels []slack.Channel, blackList []string) []slack.Channel {
	var filtered []slack.Channel
	for _, channel := range channels {
		conversation := channel.GroupConversation.Conversation
		if contains(blackList, channel.Name) {
			continue
		}
		if channel.IsGeneral {
			continue
		}
		if conversation.NumMembers > 20 {
			continue
		}
		if conversation.IsPrivate {
			continue
		}
		filtered = append(filtered, channel)
	}
	return filtered
}

func buildText(channel string) string {
	var aisatsu string
	now := time.Now()
	if 5 <= now.Hour() && now.Hour() < 12 {
		aisatsu = "おはようございます！！！"
	} else if 12 <= now.Hour() && now.Hour() < 16 {
		aisatsu = "こんにちは！！！"
	} else if 16 <= now.Hour() && now.Hour() < 23 {
		aisatsu = "こんばんは！！！"
	} else { // midnight
		aisatsu = "夜分に失礼します！！！"
	}
	return fmt.Sprintf("%s 今日のおすすめチャンネルは……これ！！！！👉👉👉👉👉 %s 👈👈👈👈👈", aisatsu, channel)
}

func parseBlackList(blackListEnv string) []string {
	return strings.Split(blackListEnv, ",")
}

func doIt(dryRun bool) {
	apiKey, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		log.Fatalln("API key not set in SLACK_TOKEN")
	}
	postChannelID, ok := os.LookupEnv("POST_CHANNEL_ID")
	if !ok {
		log.Fatalln("Post destination channel ID not set in POST_CHANNEL_ID")
	}
	blackListEnv := os.Getenv("IGNORE_CHANNELS")
	blackList := parseBlackList(blackListEnv)
	if len(blackList) > 0 {
		log.Printf("black list: %#v", blackList)
	}

	api := slack.New(apiKey)

	channels, err := api.GetChannels(true)
	if err != nil {
		log.Fatalln(err)
	}
	channels = filterChannels(channels, blackList)
	log.Printf("number of target channels is %d\n", len(channels))

	channelNames := make([]string, len(channels))
	for i, channel := range channels {
		channelName := channel.Name
		channelNames[i] = fmt.Sprintf("#%s", channelName)
	}

	todaysRecommendChannel := chooseChannel(channelNames)
	text := buildText(todaysRecommendChannel)
	log.Printf("message: %s", text)
	if !dryRun {
		log.Printf("post %s to %s", text, postChannelID)
		if _, _, err := postMessage(api, postChannelID, text); err != nil {
			log.Fatal(err)
		}
	}
}
