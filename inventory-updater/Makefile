init-project:
	go mod init ws-echo-server 

build:
	go build -o server .

run: build
	./server --addr localhost:8282 --server localhost:4222 --cluster my-stan --clientid stan-pub 

