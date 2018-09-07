#!/bin/bash
peer chaincode query --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c '{"Args":["getAllDataByID", "warming"]}' -C channel1
