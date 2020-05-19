package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/digitalocean/godo"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token = os.Getenv("DISCORD_TOKEN")
	channelId = os.Getenv("DISCORD_CHANNEL_ID")
	dropletId, _ = strconv.Atoi(os.Getenv("DROPLET_ID"))
	digitalOceanToken = os.Getenv("DIGITAL_OCEAN_TOKEN")
}

var token string
var channelId string
var dropletId int
var digitalOceanToken string

func main() {

	if token == "" {
		fmt.Println("No token provided. Please run: airhorn -t <bot token>")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("minecraft bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "minecraft")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == channelId {
		if m.Content == "!minecraft start" {
			_, err := DropletPower(true)
			if err != nil {
				_, _ = s.ChannelMessageSendTTS(channelId, "ocorreu um erro durante esta tarefa")
			}
			_, _ = s.ChannelEditComplex(channelId, &discordgo.ChannelEdit{Topic: "Minecraft server is starting"})
			_, _ = s.ChannelMessageSendTTS(channelId, "iniciando servidor")
		} else if m.Content == "!minecraft stop" {
			_, err := DropletPower(false)
			if err != nil {
				_, _ = s.ChannelMessageSendTTS(channelId, "ocorreu um erro durante esta tarefa")
			}
			_, _ = s.ChannelMessageSendTTS(channelId, "encerrando servidor")
			_, _ = s.ChannelEditComplex(channelId, &discordgo.ChannelEdit{Topic: "Minecraft server is off"})

		} else if strings.Contains(m.Content, "!minecraft") {
			_, _ = s.ChannelMessageSend(channelId, "!minecraft start => inicia o servidor")
			_, _ = s.ChannelMessageSend(channelId, "!minecraft stop => encerra o servidor")
		}
	}
	fmt.Println(m.ChannelID)
	fmt.Println(m.Author, m.Content)
}

//true: on, false off
func DropletPower(onoff bool) (*godo.Action, error) {
	client := godo.NewFromToken(digitalOceanToken)
	ctx := context.TODO()
	if onoff {
		action, _, err := client.DropletActions.PowerOn(ctx, dropletId)
		return action, err
	}
	action, _, err := client.DropletActions.PowerOff(ctx, dropletId)
	return action, err
}
