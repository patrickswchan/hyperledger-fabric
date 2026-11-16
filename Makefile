.PHONY: install up deploy demo down test

install:
	./scripts/install-fabric.sh

up:
	./scripts/up.sh

deploy:
	./scripts/deploy-chaincode.sh

demo:
	./scripts/demo.sh

down:
	./scripts/down.sh

test:
	cd chaincode/invoice && go test ./...
