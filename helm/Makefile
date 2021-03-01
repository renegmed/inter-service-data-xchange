minikube:
	minikube start --memory=4096 --driver=virtualbox

nats:
	helm install my-nats nats/nats
	helm install my-stan nats/stan --set stan.nats.url=nats://my-nats:4222 

#-------------------------------------------

portforward1: 
	kubectl port-forward web-updater-6674b5f9df-xrzf7  9090:8282

test1:
	helm template web-updater web-updater --values web-updater/updater.values.yaml 


install1:
	helm install web-updater web-updater --values web-updater/updater.values.yaml 


upgrade1:
	helm upgrade web-updater web-updater --values web-updater/updater.values.yaml 


inspect1:
	helm list
	minikube ip 
	kubectl get service/web-updater 

uninstall1:
	helm uninstall web-updater


#-------------------------------------------


test2:
	helm template web-reporter web-updater --values web-updater/reporter.values.yaml 


install2:
	helm install web-reporter web-updater --values web-updater/reporter.values.yaml 

portforward2: 
	kubectl port-forward web-reporter-6b4d7fcdd9-vjlq4 7878:7777


upgrade2:
	helm upgrade web-reporter web-updater --values web-updater/reporter.values.yaml 


inspect2:
	helm list
	minikube ip 
	kubectl get service/web-reporter 

uninstall2:
	helm uninstall web-reporter
 
