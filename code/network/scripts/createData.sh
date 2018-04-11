#!/bin/bash

for (( i = 0; i < 100; ++i ))
do
    PAYLOAD='{"Args":["createData", "1000", "test throughput data", "10", "units..", "'
    PAYLOAD=$PAYLOAD$i
    ENDING='", "marcel"]}'
    PAYLOAD=$PAYLOAD$ENDING
	2>/dev/null 1>&2 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c "${PAYLOAD}" -C channel1 &
done
wait