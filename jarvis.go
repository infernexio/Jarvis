package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// create a Guild with the name of the team and text chanal with the ip adress of the team
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
			Name:     "💚 " + ip,
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: category.ID,
		})
		if err != nil {
			fmt.Println("Error creating channel: ", err)
			os.Exit(1)
		}

		//save the channel id with the ip adress that should talk to it in a file
		data := fmt.Sprintf("%s : %s\n", ip, channel.ID)
		filename := "SOHAIL_config.txt"
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

// sets up the discord server to have the channels for the teams
func setup(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Split(m.Content, " ")[0] == "!setup" {
		args := strings.Split(m.Content, " ")[1:]
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Please provide # of teams and all ip adress of teams with X, X being team number (ex 10.X.2.2)")
			return
		} else {
			teams, err := strconv.Atoi(args[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Please provide a valid number of teams")
				return
			}

			for i := 1; i <= teams; i++ {
				CreateChannel(s, m.GuildID, strconv.Itoa(i), args[1:])
			}
		}
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

// responds with pong when ping is received
func pingpong(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" && m.ChannelID == "1084889950950015109" {
		s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

// goes through the config file and tests the connection to each channel and updates the channel name
func testConnection(s *discordgo.Session) {
	//opening the config file to get all the channel ids
	filename := "SOHAIL_config.txt"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Read the file line by line and sending ping to each channel
	for scanner.Scan() {
		line := scanner.Text()
		channelID := strings.Split(line, " : ")[1]

		//seding ping to the channel
		message, err := s.ChannelMessageSend(channelID, "ping")
		if err != nil {
			log.Fatal(err)
		}

		//waiting for 5 seconds to get the pong back
		responseReceived := false

		time.Sleep(time.Second)
		response, err := s.ChannelMessages(message.ChannelID, 1, "", message.ID, "")
		if err != nil {
			fmt.Println("Error getting next message:", err)
			return
		}
		if len(response) > 0 && response[0].Content == "pong" {
			responseReceived = true
		}

		if responseReceived == true {
			// Get the channel where the message was sent
			channel, err := s.Channel(channelID)
			if err != nil {
				panic(err)
			}

			channelEdit := &discordgo.ChannelEdit{
				Name: "💚 " + channel.Name[1:],
			}

			_, err = s.ChannelEdit(channelID, channelEdit)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// Get the channel where the message was sent
			channel, err := s.Channel(channelID)
			if err != nil {
				panic(err)
			}

			channelEdit := &discordgo.ChannelEdit{
				Name: "💔 " + channel.Name[1:],
			}

			// Set the channel name to "BAD"
			_, err = s.ChannelEdit(channelID, channelEdit)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// main function
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sess, err := discordgo.New("Bot " + os.Getenv("GITHUB_SERVER_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(pingpong)

	sess.AddHandler(setup)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer sess.Close()
	fmt.Print("Bot is running...")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	testConnection(sess)

	for {
		select {
		case <-ticker.C:
			testConnection(sess)
		}
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
