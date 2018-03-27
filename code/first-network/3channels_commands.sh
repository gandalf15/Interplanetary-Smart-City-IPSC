#!/bin/bash

echo "====== Recompiling chaincodes for channels ======"
go build --tags nopkcs11 ../chaincode/simple_chaincode/
echo "====== simple_chaincode recompiled ======"
go build --tags nopkcs11 ../chaincode/chaincode2/
echo "====== chaincode2 recompiled ======"
go build --tags nopkcs11 ../chaincode/chaincode_money/
echo "====== chaincode_money recompiled ======"

/home/marcel/fabric/bin/cryptogen generate --config=./crypto-config.yaml
export FABRIC_CFG_PATH=$PWD
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block

/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputCreateChannelTx ./channel-artifacts/channel1.tx -channelID channel1
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputCreateChannelTx ./channel-artifacts/channel2.tx -channelID channel2
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputCreateChannelTx ./channel-artifacts/channel3.tx -channelID channel3
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelOne -outputAnchorPeersUpdate ./channel-artifacts/channel1City1MSPanchors.tx -channelID channel1 -asOrg City1MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City1MSPanchors.tx -channelID channel2 -asOrg City1MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelTwo -outputAnchorPeersUpdate ./channel-artifacts/channel2City2MSPanchors.tx -channelID channel2 -asOrg City2MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City1MSPanchors.tx -channelID channel3 -asOrg City1MSP
/home/marcel/fabric/bin/configtxgen -profile TwoOrgsChannelThree -outputAnchorPeersUpdate ./channel-artifacts/channel3City2MSPanchors.tx -channelID channel3 -asOrg City2MSP
docker-compose -f docker-compose-cli.yaml -f docker-compose-couch.yaml up