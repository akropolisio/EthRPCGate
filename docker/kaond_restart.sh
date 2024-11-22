#!/bin/bash

# Check if the kaond process is running
if pgrep -x "kaond" > /dev/null
then
    echo "kaond is running, executing stop command"
    docker exec kaon_mainnet kaon-cli stop
    sleep 3 #executing too fast causes some errors
    docker restart kaon_mainnet
else
    echo "kaond is not running, restarting with -reindex"
    # Modify the restart command to include the -reindex parameter
    docker restart kaon_mainnet
    sleep 3 #executing too fast causes some errors
    docker exec kaon_mainnet kaond -reindex
fi
