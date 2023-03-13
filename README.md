# Jarvis
<img align="center" src="https://github.com/infernexio/KernelKraken/blob/main/images/Kraken.png" height="370" width="450">

## Overview
*Jarvis* is a cross-platform command and control framwork wrritten in go. It can setup a discord server to represent a competition topology.

## Setup
First add the discord bot to your discord server. If your unsure how to do that refer to the first link in the resources section.

Then run the jarvis.go with 
```
go run jarvis.go
```
or compile and run it with
```
go build jarvis.go
./jarvis.go
```

After that go to the discord server and type !setup [number of teams] [ips with X (ex. 192.168.X.12)]

Then on the target machine just run the client
```
go run client_(your os).go
```
but a better idea is to compile it so it is harder to reverse
```
go build client_(your os).go
./client_(your os)
```
And thats it. It should now the bot is running and each client will only be listening for the ipaddress that its associated to.

## Resources/References:
  https://youtu.be/XuFq7NW3ii4
  https://pkg.go.dev/github.com/bwmarrin/discordgo
