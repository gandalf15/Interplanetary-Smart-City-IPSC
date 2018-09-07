#!/bin/bash

verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to create 3 channels network ==========="
		echo
   		exit 1
	fi
}

# Ask user if recompile
function askRecompile () {
  read -p "Recompile chaincodes (y/n)? " ans
	case "$ans" in
    	y|Y )
			echo "Recompiling chaincode_data ..."
			go build --tags nopkcs11 ../chaincode/chaincode_data/
			res=$?
			verifyResult $res "chaincode_data compilation failed."
			echo "-> chaincode_data recompiled successfully."
			go build --tags nopkcs11 ../chaincode/chaincode_ad/
			res=$?
			verifyResult $res "chaincode_ad compilation failed."
			echo "-> chaincode_ad recompiled successfully."
			go build --tags nopkcs11 ../chaincode/chaincode_tokens/
			res=$?
			verifyResult $res "chaincode_tokens compilation failed."
			echo "-> chaincode_tokens recompiled successfully."
    	;;
    	n|N )
			echo "Proceeding without recompilation..."
    	;;
    	* )
    		echo "Invalid input try again"
    		askRecompile
    	;;
	esac
}

function clearContainers () {
  CONTAINER_IDS=$(docker ps -aq)
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == " " ]; then
    echo "---- No containers available for deletion ----"
  else
    docker rm -f $CONTAINER_IDS
  fi
}

function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | grep "dev\|none\|test-vp\|peer[0-9]-" | awk '{print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    echo "---- No images available for deletion ----"
  else
    docker rmi -f $DOCKER_IMAGE_IDS
  fi
}

# Tear down running network
function networkDown () {
	docker-compose -f docker-compose-cli.yaml down --volumes
	#Cleanup the chaincode containers
	clearContainers
	#Cleanup images
	removeUnwantedImages
	# remove orderer block and other channel configuration transactions and certs
	rm -rf channel-artifacts/*.block channel-artifacts/*.tx crypto-config
	# remove the docker-compose yaml file that was customized to the example
	rm -f docker-compose-e2e.yaml

}

export FABRIC_CFG_PATH=$PWD
DAEMON_MODE="0"
MODE="up"

# Parse commandline args
while getopts "m:d" opt; do
  case "$opt" in
    m)  MODE=$OPTARG
    ;;
    d)  DAEMON_MODE="1"
    ;;
  esac
done

if [ "$MODE" == "up" ]; then
	askRecompile
	~/fabric/bin/cryptogen generate --config=./crypto-config.yaml
	~/fabric/bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
	~/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputCreateChannelTx ./channel-artifacts/channel1.tx -channelID channel1
	~/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputCreateChannelTx ./channel-artifacts/channel2.tx -channelID channel2
	~/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputCreateChannelTx ./channel-artifacts/channel3.tx -channelID channel3
	~/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputAnchorPeersUpdate ./channel-artifacts/channel1City1MSPanchors.tx -channelID channel1 -asOrg City1MSP
	~/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City1MSPanchors.tx -channelID channel2 -asOrg City1MSP
	~/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City2MSPanchors.tx -channelID channel2 -asOrg City2MSP
	~/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City1MSPanchors.tx -channelID channel3 -asOrg City1MSP
	~/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City2MSPanchors.tx -channelID channel3 -asOrg City2MSP
	if [ $DAEMON_MODE == "1" ]; then
		docker-compose -f docker-compose-cli.yaml up -d
	else
		docker-compose -f docker-compose-cli.yaml up
	fi
elif [ "$MODE" == "down" ]; then
	networkDown
	rm -f ./scripts/scenarios/TxIDs.txt
	res=$?
	verifyResult $res "Cannot remove ./scripts/scenarios/TxIDs.txt"
	rm -f ./scripts/scenarios/logTxID.txt
	res=$?
	verifyResult $res "Cannot remove ./scripts/scenarios/logTxID.txt"
	rm -f ./log.txt
	res=$?
	verifyResult $res "Cannot remove /scripts/scenarios/log.txt"
else
	echo "Wrong -m arg"
fi

# docker-compose -f docker-compose-cli.yaml -f docker-compose-couch.yaml up -d


