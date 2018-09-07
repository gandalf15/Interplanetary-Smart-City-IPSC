## Interplanetary Smarty City (IPSC)


In order to run the system follow these instructions:

1. Install Go. Follow this manual: https://golang.org/doc/install
2. Install Docker. Follow this manual: https://docs.docker.com/install/
3. Install Hyperledger Fabric v 1.0 to your home directory. 
	Follow this manual: https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html
	Then follow this manual: https://hyperledger-fabric.readthedocs.io/en/latest/install.html
4. Open the root directory of the compressed file that contains the system.
5. Open directory called network
6. Execute GNU Bash script 3channels\_network\_launcher.sh

This will completely set up the network with four nodes and single ordering service.
To edit the number of nodes and orderers, edit the config files configtx.yaml and docker-compose-cli.yaml
