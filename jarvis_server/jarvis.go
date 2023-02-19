package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func CreateChannel(sess *discordgo.Session, guildID string, name string, ips []string) {
	teamName := "Team " + name
	category, err := sess.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name: teamName,
		Type: discordgo.ChannelTypeGuildCategory,
	})
	if err != nil {
		fmt.Println("Error creating category: ", err)
		os.Exit(1)
	}
	for i := 0; i < len(ips); i++ {
		ip := strings.Replace(ips[i], "X", name, -1)
		ip = strings.Replace(ip, ".", "-", -1)
		channel, err := sess.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
			Name:     ip,
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: category.ID,
		})
		if err != nil {
			fmt.Println("Error creating channel: ", err)
			os.Exit(1)
		}

		//save the channel id with the ip adress that should talk to it in a file
		data := fmt.Sprintf("%s : %s\n", ip, channel.ID)
		filename := "KernalKraken_config.txt"
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// Append data to file
		if _, err := f.Write([]byte(data)); err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}
}

func setup(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID == "1076897679151476907" {
		if strings.Split(m.Content, " ")[0] == "!setup" {
			args := strings.Split(m.Content, " ")[1:]
			if len(args) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Please provide # of teams and all ip adress of teams with X, X being team number (ex 10.X.2.2)")
				return
			} else {
				for i := 1; i < len(args); i++ {
					CreateChannel(s, m.GuildID, strconv.Itoa(i), args[1:])
				}
			}
		}
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sess, err := discordgo.New("Bot " + os.Getenv("GITHUB_SERVER_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, "pong")
		}
	})

	sess.AddHandler(setup)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer sess.Close()
	fmt.Print("Bot is running...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
