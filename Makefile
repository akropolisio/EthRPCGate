ifndef GOBIN
GOBIN := $(GOPATH)/bin
endif

# Include .env.local if it exists
ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

ifdef GATE_PORT
GATE_PORT := $(GATE_PORT)
else
GATE_PORT := 25996
endif

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GATE_DIR := "/go/src/github.com/kaonone/eth-rpc-gate"
GO_VERSION := "1.18"
ALPINE_VERSION := "3.16"
DOCKER_ACCOUNT := ripply


# Latest commit hash
GIT_SHA=$(shell git rev-parse HEAD)

# If working copy has changes, append `-local` to hash
GIT_DIFF=$(shell git diff -s --exit-code || echo "-local")
GIT_REV=$(GIT_SHA)$(GIT_DIFF)
GIT_TAG=$(shell git describe --tags 2>/dev/null)

ifeq ($(GIT_TAG),)
GIT_TAG := $(GIT_REV)
else
GIT_TAG := $(GIT_TAG)$(GIT_DIFF)
endif

check-env:
ifndef GOPATH
	$(error GOPATH is undefined)
endif

.PHONY: install
install: 
	go install \
		-ldflags "-X 'github.com/kaonone/eth-rpc-gate/pkg/params.GitSha=`./sha.sh``git diff -s --exit-code || echo \"-local\"`'" \
		github.com/kaonone/eth-rpc-gate

.PHONY: release
release: darwin linux windows

.PHONY: darwin
darwin: build-darwin-amd64 tar-gz-darwin-amd64 build-darwin-arm64 tar-gz-darwin-arm64

.PHONY: linux
linux: build-linux-386 tar-gz-linux-386 build-linux-amd64 tar-gz-linux-amd64 build-linux-arm tar-gz-linux-arm build-linux-arm64 tar-gz-linux-arm64 build-linux-ppc64 tar-gz-linux-ppc64 build-linux-ppc64le tar-gz-linux-ppc64le build-linux-mips tar-gz-linux-mips build-linux-mipsle tar-gz-linux-mipsle build-linux-riscv64 tar-gz-linux-riscv64 build-linux-s390x tar-gz-linux-s390x

.PHONY: windows
windows: build-windows-386 tar-gz-windows-386 build-windows-amd64 tar-gz-windows-amd64 build-windows-arm64 tar-gz-windows-arm64
	echo hey
#	GOOS=linux GOARCH=arm64 go build -o ./build/eth-rpc-gate-linux-arm64 github.com/kaonone/eth-rpc-gate/cli/eth-rpc-gate

docker-build-go-build:
	docker build -t kaon/go-build.ethrpcgate -f ./docker/go-build.Dockerfile --build-arg GO_VERSION=$(GO_VERSION) .

tar-gz-%:
	mv $(ROOT_DIR)/build/bin/eth-rpc-gate-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==1')-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==2') $(ROOT_DIR)/build/bin/eth-rpc-gate
	tar -czf $(ROOT_DIR)/build/eth-rpc-gate-$(GIT_TAG)-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==1' | sed s/darwin/osx/)-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==2').tar.gz $(ROOT_DIR)/build/bin/eth-rpc-gate
	mv $(ROOT_DIR)/build/bin/eth-rpc-gate $(ROOT_DIR)/build/bin/eth-rpc-gate-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==1')-$(shell echo $@ | sed s/tar-gz-// | sed 's/-/\n/' | awk 'NR==2')

# build-os-arch
build-%: docker-build-go-build
	docker run \
		--privileged \
		--rm \
		-v `pwd`/build:/build \
		-v `pwd`:$(GATE_DIR) \
		-w $(GATE_DIR) \
		-e GOOS=$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==1') \
		-e GOARCH=$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==2') \
		kaon/go-build.ethrpcgate \
			build \
			-buildvcs=false \
			-ldflags \
				"-X 'github.com/kaonone/eth-rpc-gate/pkg/params.GitSha=`./sha.sh`'" \
			-o /build/bin/eth-rpc-gate-$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==1')-$(shell echo $@ | sed s/build-// | sed 's/-/\n/' | awk 'NR==2') $(GATE_DIR)

.PHONY: quick-start
quick-start-regtest:
	cd docker && ./spin_up.regtest.sh && cd ..

.PHONY: quick-start-testnet
quick-start-testnet:
	cd docker && ./spin_up.testnet.sh && cd ..

.PHONY: quick-start-mainnet
quick-start-mainnet:
	cd docker && ./spin_up.mainnet.sh && cd ..

# docker build -t kaon/eth-rpc-gate:latest -t kaon/eth-rpc-gate:dev -t kaon/eth-rpc-gate:${GIT_TAG} -t kaon/eth-rpc-gate:${GIT_REV} --build-arg BUILDPLATFORM="$(BUILDPLATFORM)" .
.PHONY: docker-dev
docker-dev:
	docker build -t kaon/eth-rpc-gate:latest -t kaon/eth-rpc-gate:dev -t kaon/eth-rpc-gate:${GIT_TAG} -t kaon/eth-rpc-gate:${GIT_REV} --build-arg GO_VERSION=1.18 .

.PHONY: local-dev
local-dev: check-env install
	docker run --rm --name kaon_testchain -d -p 51474:51474 kaon/kaon kaond -regtest -rpcbind=0.0.0.0:51474 -rpcallowip=0.0.0.0/0 -logevents=1 -rpcuser=$(RPC_USER) -rpcpassword=$(RPC_PASSWORD) -deprecatedrpc=accounts -printtoconsole | true
	sleep 3
	docker cp ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/fill_user_account.sh kaon_testchain:.
	docker exec kaon_testchain /bin/sh -c ./fill_user_account.sh
	KAON_RPC=http://$(RPC_USER):$(RPC_PASSWORD)@localhost:51474 KAON_NETWORK=auto $(GOBIN)/eth-rpc-gate --port $(GATE_PORT) --accounts ./docker/standalone/myaccounts.txt --dev

.PHONY: local-dev-https
local-dev-https: check-env install
	docker run --rm --name kaon_testchain -d -p 51474:51474 kaon/kaon kaond -regtest -rpcbind=0.0.0.0:51474 -rpcallowip=0.0.0.0/0 -logevents=1 -rpcuser=$(RPC_USER) -rpcpassword=$(RPC_PASSWORD) -deprecatedrpc=accounts -printtoconsole | true
	sleep 3
	docker cp ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/fill_user_account.sh kaon_testchain:.
	docker exec kaon_testchain /bin/sh -c ./fill_user_account.sh > /dev/null&
	KAON_RPC=http://$(RPC_USER):$(RPC_PASSWORD)@localhost:51474 KAON_NETWORK=auto $(GOBIN)/eth-rpc-gate --port $(GATE_PORT) --accounts ./docker/standalone/myaccounts.txt --dev --https-key https/key.pem --https-cert https/cert.pem

.PHONY: local-dev-logs
local-dev-logs: check-env install
	docker run --rm --name kaon_testchain -d -p 51474:51474 kaon/kaon:dev kaond -regtest -rpcbind=0.0.0.0:51474 -rpcallowip=0.0.0.0/0 -logevents=1 -rpcuser=$(RPC_USER) -rpcpassword=$(RPC_PASSWORD) -deprecatedrpc=accounts -printtoconsole | true
	sleep 3
	docker cp ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/fill_user_account.sh kaon_testchain:.
	docker exec kaon_testchain /bin/sh -c ./fill_user_account.sh
	KAON_RPC=http://$(RPC_USER):$(RPC_PASSWORD)@localhost:51474 KAON_NETWORK=auto $(GOBIN)/eth-rpc-gate --port $(GATE_PORT) --accounts ./docker/standalone/myaccounts.txt --dev > ethrpcgate_dev_logs.txt

.PHONY: unit-tests
unit-tests: check-env
	go test -v ./... -timeout 50s

docker-build-unit-tests:
	docker build -t kaon/tests.ethrpcgate -f ./docker/unittests.Dockerfile --build-arg GO_VERSION=$(GO_VERSION) .

docker-unit-tests:
	docker run --rm -v `pwd`:/go/src/github.com/kaonone/eth-rpc-gate kaon/tests.ethrpcgate

docker-tests: docker-build-unit-tests docker-unit-tests openzeppelin-docker-compose

docker-configure-https: docker-configure-https-build
	docker/setup_self_signed_https.sh

docker-configure-https-build:
	docker build -t kaon/openssl.ethrpcgate -f ./docker/openssl.Dockerfile ./docker

# -------------------------------------------------------------------------------------------------------------------
# NOTE:
# 	The following make rules are only for local test purposes
# 
# 	Both run-ethrpcgate and run-kaon must be invoked. Invocation order may be independent, 
# 	however it's much simpler to do in the following order: 
# 		(1) make run-kaon 
# 			To stop Kaon node you should invoke: make stop-kaon
# 		(2) make run-ethrpcgate
# 			To stop eth-rpc-gate service just press Ctrl + C in the running terminal

# Runs current eth-rpc-gate implementation
run-ethrpcgate:
	@ printf "\nRunning eth-rpc-gate...\n\n"
	go run `pwd`/main.go \
		--kaon-rpc=http://$(RPC_USER):$(RPC_PASSWORD)@0.0.0.0:51474 \
		--kaon-network=auto \
		--bind=0.0.0.0 \
		--port=25996 \
		--accounts=`pwd`/docker/standalone/myaccounts.txt \
		--log-file=`pwd`/gateLogs.txt \
		--dbstring='host=127.0.0.1 port=5433 user=ethrpc password=pwd dbname=ethrpcgatedb sslmode=disable'

run-ethrpcgate-https:
	@ printf "\nRunning eth-rpc-gate...\n\n"
	go run `pwd`/main.go \
		--kaon-rpc=http://$(RPC_USER):$(RPC_PASSWORD)@0.0.0.0:51474 \
		--kaon-network=auto \
		--bind=0.0.0.0 \
		--port=25996 \
		--accounts=`pwd`/docker/standalone/myaccounts.txt \
		--log-file=`pwd`/gateLogs.txt \
		--https-cert=/etc/letsencrypt/live/testnet.kaon.one/fullchain.pem \
		--https-key=/etc/letsencrypt/live/testnet.kaon.one/privkey.pem \
		--dbstring='host=127.0.0.1 port=5433 user=ethrpc password=pwd dbname=ethrpcgatedb sslmode=disable'

# Runs docker container of kaon locally and starts kaond inside of it
run-kaon:
	@ printf "\nRunning kaon...\n\n"
	@ printf "\n(1) Starting container...\n\n"
	docker run ${kaon_container_flags} kaon/kaon kaond ${kaond_flags} > /dev/null

	@ printf "\n(2) Importing test accounts...\n\n"
	@ sleep 3
	docker cp ${shell pwd}/docker/fill_user_account.sh ${kaon_container_name}:.

	@ printf "\n(3) Filling test accounts wallets...\n\n"
	docker exec ${kaon_container_name} /bin/sh -c ./fill_user_account.sh > /dev/null
	@ printf "\n... Done\n\n"

seed-kaon:
	@ printf "\n(2) Importing test accounts...\n\n"
	docker cp ${shell pwd}/docker/fill_user_account.sh ${kaon_container_name}:.

	@ printf "\n(3) Filling test accounts wallets...\n\n"
	docker exec ${kaon_container_name} /bin/sh -c ./fill_user_account.sh
	@ printf "\n... Done\n\n"

kaon_container_name = test-chain

kaon_container_flags = \
	--rm -d \
	--name ${kaon_container_name} \
	-v ${shell pwd}/dapp:/kaon \
	-p 51474:51474

kaond_flags = \
	-testnet \
	-rpcport=51474 \
	-port=5778 \
	-logevents \
	-logtimestamps \
	-printpriority \
	-txindex \
	-txrlpindex \
	-daemon \
	-server \
	-listen \
	-txindex \
	-addrindex \
	-rpcbind=0.0.0.0:51474 \
	-rpcallowip=0.0.0.0/0 \
	-rpcuser=${RPC_USER} \
	-rpcpassword=${RPC_PASSWORD} \
	-deprecatedrpc=accounts \
	-printtoconsole
    
# Starts continuously printing Kaon container logs to the invoking terminal
follow-kaon-logs:
	@ printf "\nFollowing kaon logs...\n\n"
		docker logs -f ${kaon_container_name}

open-kaon-bash:
	@ printf "\nOpening kaon bash...\n\n"
		docker exec -it ${kaon_container_name} bash

# Stops docker container of kaon
stop-kaon:
	@ printf "\nStopping kaon...\n\n"
		docker kill `docker container ps | grep ${kaon_container_name} | cut -d ' ' -f1` > /dev/null
	@ printf "\n... Done\n\n"

restart-kaon: stop-kaon run-kaon

submodules:
	git submodules init

# Run openzeppelin tests, eth-rpc-gate/KAON needs to already be running
openzeppelin:
	cd testing && make openzeppelin

# Run openzeppelin tests in docker
# eth-rpc-gate and Kaon need to already be running
openzeppelin-docker:
	cd testing && make openzeppelin-docker

# Run openzeppelin tests in docker-compose
openzeppelin-docker-compose:
	cd testing && make openzeppelin-docker-compose
