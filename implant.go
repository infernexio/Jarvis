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
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

//gets the id of the channel that the bot should listen to
var channelID string = setup()
var configFile string = "UB_hack_config.txt"

//execute the command that is sent to the bot through the channel
func execute(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	operatingSystem := runtime.GOOS

	if operatingSystem != "windows" {
		if m.ChannelID == channelID {
			if strings.Contains(m.Content, "!") {
				command := m.Content[1:]

				cmd := exec.Command("/bin/bash", "-c", command)
				out, err := cmd.Output()
				if err != nil {
					fmt.Println(err, " on ", command)
					s.ChannelMessageSend(m.ChannelID, string("Sorry the command failed, please try again. error: "+err.Error()))
				}

				if len(out) > 2000 {
					now := time.Now()
					filename := "UB_hack_" + now.Format("2006-01-02_15-04-05.txt")
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
	} else {
		if m.ChannelID == channelID {
			if strings.Contains(m.Content, "!") {
				command := m.Content[1:]

				cmd := exec.Command("powershell", "/C", command)
				out, err := cmd.Output()
				if err != nil {
					fmt.Println(err, " on ", command)
					s.ChannelMessageSend(m.ChannelID, string("Sorry the command failed, please try again. error: "+err.Error()))
				}

				if len(out) > 2000 {
					now := time.Now()
					filename := "UB_hack_" + now.Format("2006-01-02_15-04-05.txt")
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
}

//gets the local ip address of the machine used in the setup function
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

//gets the channel id that the bot should listen to from the config file and returns it
func setup() string {
	ip, err := getLocalIP()
	if err != nil {
		fmt.Println(err)
	}

	ip = strings.Replace(ip, ".", "-", -1)

	file, err := os.Open(configFile)
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

//when it recieves ping in it's channel it responds with pong which is latter used to check if the bot is online
func checkConnection(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID == channelID {
		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, "pong")
		}
	}
}

//main function
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sess, err := discordgo.New("Bot " + os.Getenv("GITHUB_CLIENT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(checkConnection)

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
