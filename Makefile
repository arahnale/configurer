all: main remote

main:
	go build -o main ./cmd/main/main.go && chmod +x main

remote:
	go build -o remote ./cmd/remote/main.go && chmod +x remote

qa:
	./main
	# ./remote

.PHONY: all main remote qa