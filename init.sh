# setup vm

sudo apt-get update
sudo apt install golang-go -y
sudo apt install rabbitmq-server -y
sudo apt install postgresql -y
sudo pg_ctlcluster 12 main start
sudo -u postgres createdb fmm

git clone https://github.com/buidl-labs/filecoin-chain-indexer fci
cd fci && git checkout etl && go build -o main
mkdir /home/ubuntu/fci/s3data
mkdir /home/ubuntu/fci/s3data/cso
mkdir /home/ubuntu/fci/s3data/csvs
mkdir /home/ubuntu/fci/s3data/csvs/transactions
mkdir /home/ubuntu/fci/s3data/csvs/parsed_messages
