package main

import (
	"discord-lambda/commands"
	"discord-lambda/config"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

func Delete(gIdF, cIdF *string, allF *bool, disc *discordgo.Session) {
	if *cIdF != "" {
		cmd, err := disc.ApplicationCommand(config.Config.Appid, *cIdF, *gIdF)

		if err != nil {
			log.Fatal("Error fetching command: ", err)
		}

		if c := Confirm(fmt.Sprintf("You are about to delete [%s]. Do you wish to proceed? [y/n]", cmd.Name)); c != true {
			log.Fatal("Aborting")
		}

		if err := disc.ApplicationCommandDelete(config.Config.Appid, *cIdF, *gIdF); err != nil {
			log.Fatal("Error deleting command: ", err)
		}

	} else if *allF {
		c := Confirm("You are about to delete all commands.\nDo you wish to proceed? [y/n]")

		if c != true {
			log.Fatal("Aborting")
		}

		cmds, err := disc.ApplicationCommands(config.Config.Appid, *gIdF)

		if err != nil {
			log.Fatal("Error fetching commands")
		}

		for j := range cmds {
			if err := disc.ApplicationCommandDelete(config.Config.Appid, cmds[j].ID, *gIdF); err != nil {
				log.Fatalf("Error deleting command [%s]: %s", cmds[j].Name, err)
			} else {
				fmt.Printf("Deleted %s\n", cmds[j].Name)
			}
		}

	} else {
		log.Fatal("Please specify a command with -commandid {ID} or -all to delete all guild commands")
	}

}

func Confirm(prompt string) bool {
	fmt.Print(prompt)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatalf("Error reading from stdin: %s\n", err)
	}
	var (
		Continue = []string{"y", "Y", "yes", "Yes", "YES"}
		Stop     = []string{"n", "N", "no", "No", "NO"}
	)

	for i := 0; i < len(Continue); i += 1 {
		if response == Continue[i] {
			return true
		} else if response == Stop[i] {
			return false
		}
	}

	fmt.Println("Error! Response must be one of ", Continue, Stop, " Please try again.")

	return Confirm(prompt)
}

var actions = []string{
	"list",
	"delete",
	"show",
	"upload",
}

func main() {

	var (
		allF = flag.Bool("all", false, "Whether to apply the operation to all entities")
		cIdF = flag.String("commandid", "", "ID of the relevant command")
		gIdF = flag.String("guildid", "", "ID of the relevant guild")
	)

	flag.Parse()

	//fmt.Printf("All = %v\nCommand ID = %s\nGuild ID = %s\n", *allF, *cIdF, *gIdF)

	disc, err := discordgo.New("Bot " + config.Config.Bottoken)

	if err != nil {
		log.Fatalf("Error creating discord client: %s\n", err.Error())
	}

	err = disc.Open()

	if err != nil {
		log.Fatalf("Error opening a connection to discord: %s\n", err.Error())
	}
	
	//jsonb, _ := json.MarshalIndent(&disc.State.User, "", "\t")

	//fmt.Println("Acting as:",string(jsonb))

	action := flag.Arg(0)

	if action == "" {
		log.Fatal("Error! Please supply an action (one of ", actions, ")")
	}

	isValid := func() bool {
		for i := range actions {
			if action == actions[i] {
				return true
			}
		}
		return false
	}()

	if !isValid {
		log.Fatal("Error! action must be one of ", actions)
	}

	switch action {
	case "list":
		cmds, err := disc.ApplicationCommands(config.Config.Appid, *gIdF)

		if err != nil {
			log.Fatal("Error fetching commands: ", err)
		}

		for j := range cmds {
			jsonb, err := json.MarshalIndent(&cmds[j], "", "\t")
			if err != nil {
				log.Fatal("Error marshaling response: ", err)
			}
			fmt.Println(string(jsonb))
		}

	case "delete":
		Delete(gIdF, cIdF, allF, disc)
	case "show":
		if *cIdF != "" {
			cmd, err := disc.ApplicationCommand(config.Config.Appid, *cIdF, *gIdF)
			if err != nil {
				log.Fatal("Error fetching command: ", err)
			}
			jsonb, err := json.MarshalIndent(&cmd, "", "\t")
			if err != nil {
				log.Fatal("Error marshaling json: ", err)
			}
			fmt.Println(string(jsonb))
		} else {
			log.Fatal("Please provide a command id with -commandid {ID}")
		}
	case "upload":
		fmt.Println("Note: If your command already exists it will be updated")
		for j := range commands.Commands {
			cmd := commands.Commands[j].Command
			_, err := disc.ApplicationCommandCreate(config.Config.Appid, *gIdF, &cmd)
			if err != nil {
				log.Fatalf("Error uploading command [%s]: %s", cmd.Name, err)
			} else {
				fmt.Printf("Uploaded command [%s]\n", cmd.Name)
			}
		}
	}

}
