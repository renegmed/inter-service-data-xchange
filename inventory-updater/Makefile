init-project:
	go mod init inventory-updater

build:
	go build -o updater .

run: build
	./updater --addr localhost:8282 --server localhost:4222 --cluster my-stan --clientid stan-pub 

