init-project:
	go mod init inventory-reporter 

docker:

tohub:


build:
	go build -o report .

run: build
	./report --addr localhost:7676 --server localhost:4222 --cluster my-stan --clientid stan-sub


run2: 
	./report --addr localhost:7777 --server localhost:4222 --cluster my-stan --clientid stan-sub2
