CC=go

APP=lcongra

all: $(APP)

go.mod:
	go mod tidy

lcongra: go.mod
	$(CC) build -o build/lcongra cmd/lcongra/*.go 
