package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

const TOKEN = ""

func interactionHandle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		MinesweeperCommand(s, i)
	} else {
		MineButtonHandle(s, i)
	}
}

func main() {
	s, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatal("Error creating bot: ", err)
	}
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	s.AddHandler(interactionHandle)
	_, err = s.ApplicationCommandCreate("806165643040784414", "", &discordgo.ApplicationCommand{
		Name:        "minesweeper",
		Description: "a minesweeper game",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}
