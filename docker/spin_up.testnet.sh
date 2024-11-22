#!/bin/sh
docker-compose -f ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/quick_start/docker-compose.testnet.yml up -d
sleep 3 #executing too fast causes some errors
docker cp ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/fill_user_account.sh kaon_testchain:.
docker exec kaon_testnet /bin/sh -c ./fill_user_account.sh