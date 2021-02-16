package main

import (
	"crypto/ed25519"
	"discord-lambda/commands"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"discord-lambda/config"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	lambdaSdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/bwmarrin/discordgo"
)

func handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	//log.Print(e)

	pubkey_b, err := hex.DecodeString(config.Config.Publickey)
	if err != nil {
		return events.APIGatewayProxyResponse{}, errors.New("Couldn't decode the public key")
	}
	if e.Body == "" {
		log.Print("400 No data")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error":"No body data"}`,
		}, nil
	}

	var body []byte

	if e.IsBase64Encoded {
		body_b, err := base64.StdEncoding.DecodeString(e.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{}, errors.New(fmt.Sprintf("Couldn't decode request body [%s]: %s", body, err))
		}
		body = body_b
	} else {
		body = []byte(e.Body)
	}

	//fmt.Println(body)

	pubkey := ed25519.PublicKey(pubkey_b)

	XSig, ok := e.Headers["x-signature-ed25519"]

	if !ok {
		log.Print("400 No Signature header")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Missing 'X-Signature-Ed25519' header"}`,
		}, nil
	}

	XSigTime, ok := e.Headers["x-signature-timestamp"]

	if !ok {
		log.Print("400 Missing Timestamp header")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Missing 'X-Signature-Timestamp' header"}`,
		}, nil
	}

	XSigB, err := hex.DecodeString(XSig)

	if err != nil {
		return events.APIGatewayProxyResponse{}, errors.New("Couldn't decode signature")
	}

	SignedData := []byte(XSigTime + string(body))

	if !ed25519.Verify(pubkey, SignedData, XSigB) {
		log.Print("401 Unauthorized")
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
		}, nil
	} else {
		//authorized
		var inter discordgo.Interaction
		err := json.Unmarshal(body, &inter)

		if err != nil {
			log.Printf("Error decoding interaction: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
			}, nil
		}

		switch {
		case inter.Type == 1:
			{
				log.Print("200 Type 1 Ping")
				return events.APIGatewayProxyResponse{
					StatusCode: 200,
					Body:       `{"type":1}`,
				}, nil
			}
		case inter.Type == 2:
			{
				log.Printf("Application command [%s]", inter.Data.Name)
				log.Print(string(body))
				handler, ok := commands.Commands[inter.Data.Name]
				if !ok {
					return events.APIGatewayProxyResponse{
						StatusCode: 404,
						Body:       `{"error": "Command not found"}`,
					}, nil
				}

				response, cont, err := handler.Handler(&inter)

				if err != nil {
					log.Printf("Error in handler: %s", err)
					return events.APIGatewayProxyResponse{}, err
				} else {
					res_body, err := json.Marshal(&response)

					if err != nil {
						return events.APIGatewayProxyResponse{}, err
					}

					if cont {
						log.Print("Attempting to invoke task lambda")
						session, err := session.NewSession()
						if err == nil {
							client := lambdaSdk.New(session)

							input := &lambdaSdk.InvokeInput{
								FunctionName:   aws.String(config.Config.Tasklambda),
								InvocationType: aws.String("Event"),
								Payload:        body,
							}

							_, err := client.Invoke(input)

							if err != nil {
								log.Print("Failed to invoke lambda with error: ", err)
								return events.APIGatewayProxyResponse{}, err
							}

						} else {
							log.Printf("Error creating continuation session")
							return events.APIGatewayProxyResponse{}, errors.New("Error creating continuation session")
						}
					}

					return events.APIGatewayProxyResponse{
						StatusCode: 200,
						Body:       string(res_body),
					}, nil
				}
			}
		default:
			{
				return events.APIGatewayProxyResponse{
					StatusCode: 501,
				}, nil
			}
		}
	}
}

func main() {
	lambda.Start(handle)
}
