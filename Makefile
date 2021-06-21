CC=go

APP=lcongra

all: $(APP) term

go.mod:
	go mod tidy

lcongra: go.mod
	$(CC) build -o build/lcongra cmd/lcongra/*.go 

term:  
	$(CC) build -o build/term client/term/*.go
