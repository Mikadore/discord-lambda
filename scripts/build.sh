export GOOS=linux
export CGO_ENABLED=0

mkdir -p out

go build -o main    lambda-endpoint/main.go 
zip out/endp.zip main

go build -o main    lambda-task/main.go 
zip out/task.zip main

rm main