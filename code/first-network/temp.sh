## Install chaincode on Peer0/City1 and Peer2/City2
echo "Installing simple_chaincode on Peer0/City1..."
installChaincode 0 simple_chaincode github.com/hyperledger/fabric/chaincode/simple_chaincode
# echo "Installing simple_chaincode on Peer1/City1..."
# installChaincode 1 simple_chaincode github.com/hyperledger/fabric/chaincode/simple_chaincode

echo "Installing chaincode2 on Peer0/City1..."
installChaincode 0 chaincode2 github.com/hyperledger/fabric/chaincode/chaincode2
# echo "Install chaincode2 on Peer1/City1..."
# installChaincode 1 chaincode2 github.com/hyperledger/fabric/chaincode/chaincode2
echo "Install chaincode2 on Peer2/City2..."
installChaincode 2 chaincode2 github.com/hyperledger/fabric/chaincode/chaincode2
# echo "Install chaincode2 on Peer3/City2..."
# installChaincode 3 chaincode2 github.com/hyperledger/fabric/chaincode/chaincode2

echo "Installing chaincode_tokens on Peer0/City1..."
installChaincode 0 chaincode_tokens github.com/hyperledger/fabric/chaincode/chaincode_tokens
# echo "Install chaincode_tokens on Peer1/City1..."
# installChaincode 1 chaincode_tokens github.com/hyperledger/fabric/chaincode/chaincode_tokens
echo "Installing chaincode_tokens on Peer2/City2..."
installChaincode 2 chaincode_tokens github.com/hyperledger/fabric/chaincode/chaincode_tokens
# echo "Install chaincode_tokens on Peer3/City2..."
# installChaincode 3 chaincode_tokens github.com/hyperledger/fabric/chaincode/chaincode_tokens

#Instantiate chaincode on Peer0/City1
echo "Instantiating simple_chaincode on Peer0/City1..."
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
PAYLOAD='{"Args":["createDataEntry", "1", "test data", "50", "Celsius", "20180321163750", "marcel"]}'
POLICY="AND ('City1MSP.member')"
instantiateChaincode 0 simple_chaincode

echo "Instantiating chaincode2 on Peer2/City2..."
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["createDataEntry", "1", "test data", "50", "Celsius", "20180321163750", "marcel"]}'
POLICY="OR ('City1MSP.member','City2MSP.member')"
instantiateChaincode 2 chaincode2

echo "Instantiating chaincode_tokens on Peer2/City2..."
CHANNEL_NAME="${CHANNEL_NAME_BASE}3"
PAYLOAD='{"Args":["2", "1000"]}'
POLICY="OR ('City1MSP.member','City2MSP.member')"
instantiateChaincode 2 chaincode_tokens

# Invoke on chaincode on Peer0/City1
echo "Sending invoke transaction on City1/peer0 on simple_chaincode ..."
sleep 2
CHANNEL_NAME="${CHANNEL_NAME_BASE}1"
PAYLOAD='{"Args":["createDataEntry", "1", "test data", "50", "Celsius", "20180321163750", "marcel"]}'
chaincodeInvoke 0 simple_chaincode

# Invoke on chaincode on Peer2/City2
echo "Sending invoke transaction on Peer2/City2 on chaincode2 ..."
sleep 2
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["createDataEntry", "1", "test data", "???", "???", "20180321163750", "marcel"]}'
chaincodeInvoke 0 chaincode2

# Invoke on chaincode on Peer2/City2
echo "Sending invoke transaction on Peer2/City2 on chaincode2 revealDataValue"
sleep 2
CHANNEL_NAME="${CHANNEL_NAME_BASE}2"
PAYLOAD='{"Args":["revealDataValue", "simple_chaincode", "1", "channel1"]}'
chaincodeInvoke 0 chaincode2