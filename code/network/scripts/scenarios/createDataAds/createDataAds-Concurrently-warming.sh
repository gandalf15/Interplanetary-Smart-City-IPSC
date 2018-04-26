#!/bin/bash

ENDING='", "marcel", "1", "2"]}'

for (( i = 100001; i < 100101; ++i )); do
	PAYLOAD='{"Args":["createDataEntryAd", "warming", "warming up", "???", "Celsius", "'
        PAYLOAD=$PAYLOAD$i
        PAYLOAD=$PAYLOAD$ENDING
        2>/dev/null 1>&2 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_ad -c "${PAYLOAD}" -C channel2
done
