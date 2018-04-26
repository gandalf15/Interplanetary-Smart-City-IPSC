#!/bin/bash


ENDING='", "marcel"]}'

for (( i = 100001; i < 100101; ++i )); do
	PAYLOAD='{"Args":["createData", "warming", "warming up", "10", "warm", "'
        PAYLOAD=$PAYLOAD$i
        PAYLOAD=$PAYLOAD$ENDING
	2>/dev/null 1>&2 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c "${PAYLOAD}" -C channel1
done
