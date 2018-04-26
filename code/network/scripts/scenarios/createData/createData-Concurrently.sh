#!/bin/bash

ENDING='", "marcel"]}'

for (( i = 0; i < 100; ++i )); do
    for (( j = 0; j < 10; ++j )); do
        PAYLOAD='{"Args":["createData", "throughput", "test throughput data", "10", "units..", "'
        PAYLOAD=$PAYLOAD$i$j
        PAYLOAD=$PAYLOAD$ENDING
        2>/dev/null 1>&2 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c "${PAYLOAD}" -C channel1 &
				#peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c "${PAYLOAD}" -C channel1 &
   done
    wait
done
