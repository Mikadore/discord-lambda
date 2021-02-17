# Discord slash commands in AWS Lambda
## Rationale
Serverless is becoming more and more popular.
This kind of infrastructure is not only very budget friendly,
but also hugely scalable and once set up requiring almost 0 
maintenance. Recently, discord has introduced 
a new kind of feature, interactions. The only kind of 
interactions at the moment are slash commands. Interactions can be
configured to act as post requests to custom endpoints.
This repository is a template for AWS Lambda functions serving
as endpoints for discord to use.
## Downsides
So far I've mostly only mentioned the upsides of a serverless approach.
It's worth considering the downsides of this approach as well, amongst them are
- Latency. Cold starts are performance killers & some API limitations force us to employ less than optimal solutions.
- Versatility. Slash commands essentially only provide us with some input data and a context to send webhooks. Many features a conventional bot might implement are therefore unavailable. 
- Working with AWS services in general can be a bit tedious
# Setup
In this section we'll be creating 2 lambda functions. 
To understand why that is necessary, please read the Usage section.

## Discordgo
This template uses discordgo for its types & some AWS sdk
libraries. As of today, 2021-02-16, discordgo hasn't merged 
support for slash commands yet. 
As I cannot get Go's packages to work, I've done some black magic 
and merged the relevant branch myself locally.
```sh
cd $(go env GOPATH)/src/github.com/bwmarrin/discordgo
git remote rm origin
git remote add origin https://github.com/FedorLap2006/discordgo
git pull origin slashes
```
In the future just using discordgo normally should suffice.
## Building
To build our lambda code, we just invoke
```
go build lambda-task/main.go
go build lambda-endpoint/main.go
```
To use them in lambda, we need to 
- set some enviromental variables, e.g. `GOOS=linux` and `CGO_ENABLED=0`
- zip the binaries

The bash script `build.sh` does that for us and conveniently places the resulting zip files in an `out` folder.

## AWS
These are the steps which can be used to get lambda running. If you're already experienced with lambda, you can change these as you see fit.
I'm using the aws cli for clarity & efficiency's sake.

First we create a policy for our first lambda. Pay attention to the ARN we get back. We're gonna need it.
```
aws iam create-policy --policy-name LambdaInvokeAll --policy-document file://trust-policy.json
```
Now, we create an execution role for our first and second lambda function. We are again going to need the returned ARNs
```
aws iam create-role \
--role-name DiscordEndpointRole \
--assume-role-policy-document file://trust-policy.json

aws iam create-role \
--role-name DiscordTaskRole \
--assume-role-policy-document file://trust-policy.json
```

And attach the needed policies
```
aws iam attach-role-policy \
--role-name DiscordTaskRole \
--policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

aws iam attach-role-policy \
--role-name DiscordEndpointRole \
--policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

aws iam attach-role-policy \
--role-name DiscordEndpointRole \
--policy-arn {OUR INVOCATION ARN}
```
Lastly, we create the actual lambda functions.
```
aws lambda create-function --function-name DiscordEndpoint \
--handler main --runtime go1.x \
--role {Endpoint Role ARN} \
--zip-file fileb://out/endp.zip

aws lambda create-function --function-name DiscordTask \
--handler main --runtime go1.x \
--role {Task Role ARN} \
--zip-file fileb://out/task.zip \
--timeout 900
```

**We also need to add a trigger to the endpoint lambda, there's probably a complicated way to do this from the cli, but I've just used the web interface.
Make sure to choose API Gateway -> Create an API -> HTTP API and choose Security=Open**

Now, under `details` you'll see the API endpoint. Copy that link into discord.
# Configuration
We now need to set all the correct values in `config/config.go`. 
`Appid` and `Publickey` can both be found in the discord 
overview of our application. 
For the `Bottoken` we need to create a bot and use its token.
For the `Tasklambda` value use the name of our DiscordTask function.

After recompiling, to update our code we can use these commands (Note: these do take a few seconds to update):
```
aws lambda update-function-code --function-name DiscordEndpoint \
--zip-file fileb://out/endp.zip 

aws lambda update-function-code --function-name DiscordTask \ 
--zip-file fileb://out/task.zip 
```
Which I've put into `update.sh` as well. (Note: `update.sh` also builds the lambda code and *cli*, more on that later)

# Usage
After you've set up AWS (by either following my setup instructions or doing your own thing),
you can start developing & testing  your application.
## Program structure
All discord specific types are supplied by discordgo.
(Or rather [PR 856](https://github.com/bwmarrin/discordgo/pull/856) which hasn't been merged as of now - see the Setup)

The directory `commands`, or equivalently the package, contains
a `Commands` map, which maps a string (the command name)
to a `Command` struct as defined bellow
```go
type Command struct {
	Command      discordgo.ApplicationCommand
	Handler      HandlerSig
	Continuation ContinuationSig
}
```
To add a command add an entry to that map. For example:
```go
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
	Handler: AckSourceContinue,
	Continuation: Timer,
},
```
On import, the module checks that each key matches the Name field
in the accompanying `ApplicationCommand` struct.
The `Handler` field is a function with the signature 
```go
func(*discordgo.Interaction) (discordgo.InteractionResponse, bool, error)
```
The handler not being `nil` is checked during import.
The input is the interaction received from discord. It returns an
interaction response, a boolean signifying whether to continue
(more below) and an error. If the error is not `nil` the server 
will return `500` (Everything is logged in Lambda - including the error).

The Discord API requires us to respond within 3 seconds of receiving the interaction.
Any tasks longer than that are *delegated to a second lambda*.
If your initial handler returns `true` for its boolean return value,
the endpoint lambda will invoke the second task lambda (using its name as defined in the config struct (`config/config.go`)).
The lambda is simply passed the interaction struct. It's not 
possible to pass other custom data to it (you're welcome to modify the code and/or issue a PR). 
The task lambda then invokes the `Continuation` field. If it's nil 
the lambda will fail, logging any errors. (But since this separate from the HTTP endpoint lambda failing does not affect discord)
The signature for the continuation handler is:
```go
func(*discordgo.Interaction) error
```
If error isn't nil the lambda will fail, logging the error.

To add custom commands you need to
1. Write a handler (and optionally a continuation handler)
2. Add the command to the `Commands` map
3. Recompile and update the code of both lambdas
4. **Update the commands on discords side - see the CLI**

Note that since your lambdas are standalone and each command is equal to an 
invocation, you can automate the testing of your commands pretty easily.

Also note that slash commands by themselves only actually allow us to receive commands and send responses.
AWS Lambda has a hard limit for runtime at 15 minutes, making many tasks
such as playing music hugely infeasible. On the other hand, you obviously can interact with discord in Lambda using discordgo, REST API calls, or any other way you see fit. (Do note that for many discordgo operations you conditionally need to have called the `Open` function for them to work)

## CLI
To actually use slash commands, you must update them on discords side.
Obviously, this *could* be done using curl, python, etc. However since all
of your commands are already defined in `commands/commands.go`, I have added a CLI to automate some tasks. 
(The code is pretty messy & could be improved upon, although it suffices at its job, feel free to open an issue or PR)

**Note: to upload your commands you need to recompile the CLI**

There are currently 4 actions the CLI can perform
- list
- delete
- show
- upload

The CLI also takes in these flags:
- -commandid: Specifies the command's ID where applicable
- -guildid: Specifies the guild's ID where applicable
- -all: Specifies for an operation to be performed on all commands

### Example uses:
List all guild commands:
```
./slashes -guildid={ID} list   
```
Show a command:
```
./slashes -commandid={ID} show
```
Delete all global commands:
```
./slashes -all delete
```

Upload your commands globally: (Upload will create a command anew OR modify an already existing one if the names match):
```
./slashes -guildid=784518984947073025 upload
```
To upload globally just omit the `-guildid` flag.

**Note: You need to have `config/config.go` configured correctly for the CLI to work**

**Note: order matters. `./slashes -guildid={ID} list` will word whereas `./slashes list -guildid={ID}` won't**