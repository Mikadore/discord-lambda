package commands

import (
	"github.com/bwmarrin/discordgo"
)

func Ping(interaction *discordgo.Interaction) (discordgo.InteractionResponse, bool, error) {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: "Pong!",
		},
	}, false, nil
}
