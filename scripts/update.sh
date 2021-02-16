bash scripts/build.sh
aws lambda update-function-code --function-name DiscordEndpoint --zip-file fileb://out/endp.zip 

aws lambda update-function-code --function-name DiscordTask --zip-file fileb://out/task.zip 