#!/bin/bash

ENDING='", "marcel", "1", "2"]}'
for (( i = 0; i < 200; ++i )); do
    for (( j = 0; j < 5; ++j )); do
        PAYLOAD='{"Args":["createDataEntryAd", "1000", "test throughput data", "???", "Celsius", "'
        PAYLOAD=$PAYLOAD$i$j
        PAYLOAD=$PAYLOAD$ENDING
        2>/dev/null 1>&2 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_ad -c "${PAYLOAD}" -C channel2 &
    done
    wait
done

# peer chaincode query --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_ad -c '{"Args":["getAllDataAdByID", "1000"]}' -C channel2