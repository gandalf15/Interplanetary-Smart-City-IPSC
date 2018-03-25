#!/bin/bash
# This commands are executed inside CLI container immediately when started
# It creates blocks for channels 1,2,3 and then it broadcasts them to orderer and connect the channels
peer channel create -o orderer.zak.codes:7050 -c channel1 -f ./channel-artifacts/channel1.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
sleep 3
peer channel create -o orderer.zak.codes:7050 -c channel2 -f ./channel-artifacts/channel2.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
sleep 3
peer channel create -o orderer.zak.codes:7050 -c channel3 -f ./channel-artifacts/channel3.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem
sleep 3
peer channel join -b channel1.block
sleep 3
peer channel join -b channel2.block
sleep 3
peer channel join -b channel3.block
sleep 3