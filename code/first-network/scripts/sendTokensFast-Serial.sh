#!/bin/bash

chaincodeInvoke () {
	peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_tokens -c "${PAYLOAD}" -C channel3 >&logTxID.txt
	# Extract the returned TxID
    RETURNED_TXID=$(cat logTxID.txt | awk -F"payload:" '{print $2}')
    RETURNED_TXID=$(echo $RETURNED_TXID | awk -F">" '{print $1}')
    echo $RETURNED_TXID >> TxIDs.txt
}

for (( i = 0; i < 100; ++i ))
do
    PAYLOAD='{"Args":["sendTokensFast", "1", "2", "1", "false"]}'
    # Run the function in subshells
	chaincodeInvoke
done