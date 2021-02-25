install:
	helm install my-nats nats/nats
	helm install my-stan nats/stan --set stan.nats.url=nats://my-nats:4222

portforward:
	kubectl port-forward my-nats-0 4222:4222
