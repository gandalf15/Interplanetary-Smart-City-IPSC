#!/bin/bash
peer chaincode query --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_ad -c '{"Args":["getAllDataAdByID", "throughput"]}' -C channel2
