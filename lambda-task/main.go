package main

import (
	"discord-lambda/commands"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"log"
)

func handler(interaction discordgo.Interaction) error {
	jsonb, _ := json.Marshal(&interaction)

	log.Print("Received interaction: ", string(jsonb))

	cmd, ok := commands.Commands[interaction.Data.Name]

	if !ok {
		return errors.New("404 Command not found")
	}

	if cmd.Continuation == nil {
		return errors.New("500 Continuation is nil")
	}

	return cmd.Continuation(&interaction)
}

func main() {
	lambda.Start(handler)
}
