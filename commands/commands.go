package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type (
	HandlerSig      = func(*discordgo.Interaction) (discordgo.InteractionResponse, bool, error)
	ContinuationSig = func(*discordgo.Interaction) error
)

//checks for inconsistencies in Commands
func init() {
	for key := range Commands {
		if Commands[key].Command.Name != key {
			log.Fatalf("Error: Key [%s] in Commands doesn't equal the Name property [%s]\n", key, Commands[key].Command.Name)
		}
		if Commands[key].Handler == nil {
			log.Fatalf("Error: Handler for command [%s] is nil \n", key)
		}
	}
}

//Use these to immediately continue to the task lambda
func AckContinue(*discordgo.Interaction) (discordgo.InteractionResponse, bool, error) {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseAcknowledge,
	}, true, nil
}
func AckSourceContinue(*discordgo.Interaction) (discordgo.InteractionResponse, bool, error) {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseACKWithSource,
	}, true, nil
}

type Command struct {
	Command      discordgo.ApplicationCommand
	Handler      HandlerSig
	Continuation ContinuationSig
}

var Commands = map[string]Command{
	"ping": {
		Command: discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Pong!",
		},
		Handler: Ping,
	},
	"timer": {
		Command: discordgo.ApplicationCommand{
			Name:        "timer",
			Description: "Sets a timer. Tick tock, tick tock...",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "Seconds",
					Description: "How many seconds to set the timer for",
					Required:    true,
				},
			},
		},
		Handler:      AckSourceContinue,
		Continuation: Timer,
	},
}
