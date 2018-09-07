#!/bin/bash



# Parse commandline args
while getopts "b:e:" opt; do
  case "$opt" in
    b)  BEGIN_AT=$OPTARG
    ;;
    e)  END_AT=$OPTARG
    ;;
  esac
done

ENDING='","1"]}'
NUM=100000000000
for (( i = BEGIN_AT; i < END_AT; ++i )); do
    for (( j = 0; j < 10; ++j )); do
        PAYLOAD='{"Args":["createData","1","1","1","1","'
        PAYLOAD=$PAYLOAD$((NUM+($i$j)))
        PAYLOAD=$PAYLOAD$ENDING
	      #echo $PAYLOAD >> payload_output
	    2>/dev/null 1>&2 peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c "${PAYLOAD}" -C channel1 &
      #peer chaincode invoke --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/zak.codes/orderers/orderer.zak.codes/msp/tlscacerts/tlsca.zak.codes-cert.pem -n chaincode_data -c "${PAYLOAD}" -C channel1 &
    done
    wait
done

