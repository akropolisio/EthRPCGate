#!/bin/sh
docker-compose -f ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/quick_start/docker-compose.mainnet.yml up -d
sleep 3 #executing too fast causes some errors
docker cp ${GOPATH}/src/github.com/kaonone/eth-rpc-gate/docker/fill_user_account.sh kaon_testchain:.
docker exec kaon_mainnet /bin/sh -c ./fill_user_account.sh