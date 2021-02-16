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
[W.I.P]