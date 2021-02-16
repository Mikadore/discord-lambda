package commands

import (
	//"errors"
	//"fmt"
	//"time"
	"discord-lambda/config"
	"time"
	"github.com/bwmarrin/discordgo"
)
//Note, this is certainly not a good way to implement a timer
//it's just a simple example for how continuation works
func Timer(interaction *discordgo.Interaction) error {
	discord, err := discordgo.New()
	if err != nil {
		return err
	}
	time.Sleep(time.Second * time.Duration(interaction.Data.Options[0].Value.(float64)))

	_, err = discord.FollowupMessageCreate(config.Config.Appid, interaction, true, &discordgo.WebhookParams{
		Content: "Rrrrring! Time's up!",
	})
	
	return err
}
