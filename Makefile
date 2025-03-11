build:
	docker build -t sefiacr.azurecr.io/dns:local .
	docker push sefiacr.azurecr.io/dns:local
	kind load docker-image sefiacr.azurecr.io/dns:local

run:
	kubectl create -f role.yaml,rolebinding.yaml,dns.yaml

clean:
	kubectl delete po dns-test-pod --force --grace-period=0

print:
	kubectl logs dns-test-pod

all: clean build run print