img_tag = "username/kaonportal:testethrpcgate"
container_name = "kaonportal-testethrpcgate"

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -o ./docker/eth-rpc-gate-linux-amd64 github.com/kaonone/eth-rpc-gate/cli/eth-rpc-gate
	docker build ./docker -t $(img_tag)

.PHONY: run
run:
	cd .. && make run-ethrpcgate-https

.PHONY: stop
stop:
	docker stop $(container_name)

.PHONY: exec
exec:
	docker exec -it $(container_name) sh

.PHONY: init-container
init-container:
	docker exec -it kaonportal-testethrpcgate sh init-container.sh

.PHONY: logs
logs:
	docker logs --tail 1000  -f $(container_name)

.PHONY: clean-kaon-data
clean-kaon-data:
	rm -rf ./docker/.kaon

.PHONY: deploy-contracts
deploy-contracts:
	cd pet-shop-tutorial && truffle migrate --reset
	cp pet-shop-tutorial/build/contracts/*.json pet-shop-tutorial/js

.PHONY: serve-pet-shop
serve-pet-shop:
	cd pet-shop-tutorial; yarn
	echo "Two prefund accounts you can login with: \n 0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead \n 0x3f501c368cb9ddb5f27ed72ac0d602724adfa175"
	cd pet-shop-tutorial; yarn dev &> /dev/null
