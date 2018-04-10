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

# Ask user if generate new crypto
function askGenerateCrypto () {
  read -p "Generate new Crypto files and artifacts (y/n)? " ans
	case "$ans" in
    	y|Y )
			echo "Generating new crypto files and artifacts ..."
			/home/marcel/fabric/bin/cryptogen generate --config=./crypto-config.yaml
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputCreateChannelTx ./channel-artifacts/channel1.tx -channelID channel1
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputCreateChannelTx ./channel-artifacts/channel2.tx -channelID channel2
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputCreateChannelTx ./channel-artifacts/channel3.tx -channelID channel3
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputAnchorPeersUpdate ./channel-artifacts/channel1City1MSPanchors.tx -channelID channel1 -asOrg City1MSP
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City1MSPanchors.tx -channelID channel2 -asOrg City1MSP
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City2MSPanchors.tx -channelID channel2 -asOrg City2MSP
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City1MSPanchors.tx -channelID channel3 -asOrg City1MSP
			/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City2MSPanchors.tx -channelID channel3 -asOrg City2MSP
			echo "Crypto files and artifacts generated"
    	;;
    	n|N )
			echo "Proceeding without new Crypto files and artifacts..."
    	;;
    	* )
    		echo "Invalid input try again"
    		askGenerateCrypto
    	;;
	esac
}

export FABRIC_CFG_PATH=$PWD
askRecompile
# askGenerateCrypto
/home/marcel/fabric/bin/cryptogen generate --config=./crypto-config.yaml
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputCreateChannelTx ./channel-artifacts/channel1.tx -channelID channel1
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputCreateChannelTx ./channel-artifacts/channel2.tx -channelID channel2
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputCreateChannelTx ./channel-artifacts/channel3.tx -channelID channel3
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputAnchorPeersUpdate ./channel-artifacts/channel1City1MSPanchors.tx -channelID channel1 -asOrg City1MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City1MSPanchors.tx -channelID channel2 -asOrg City1MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City2MSPanchors.tx -channelID channel2 -asOrg City2MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City1MSPanchors.tx -channelID channel3 -asOrg City1MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City2MSPanchors.tx -channelID channel3 -asOrg City2MSP
# docker-compose -f docker-compose-cli.yaml -f docker-compose-couch.yaml up -d
docker-compose -f docker-compose-cli.yaml up -d