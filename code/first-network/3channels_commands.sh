#!/bin/bash
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
# docker exec -it cli bash
# peer channel create -o orderer.zak.codes:7050 -c channel1 -f ./channel-artifacts/channel1.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
# peer channel create -o orderer.zak.codes:7050 -c channel2 -f ./channel-artifacts/channel2.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
# peer channel create -o orderer.zak.codes:7050 -c channel3 -f ./channel-artifacts/channel3.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
# peer channel join -b channel1.block
# peer channel join -b channel2.block
# peer channel join -b channel3.block