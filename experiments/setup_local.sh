sudo apt update
# sudo apt-get -y upgrade

## setup go
wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz
sudo tar -xvf go1.13.3.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:$PATH

## setup python, pip and its library.

## download data
mkdir /proj/yarnrm-GP0/google_data_trace
cd /proj/yarnrm-GP0/google_data_trace
wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1mh3eWQUr0_Y8fkBqiZydN186NgwSJQQh' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1mh3eWQUr0_Y8fkBqiZydN186NgwSJQQh" \
  -O tasks.tar && rm -rf /tmp/cookies.txt
sudo tar -xvf tasks.tar
mv tasks-new tasks

wget --load-cookies /tmp/cookies.txt \
  "https://docs.google.com/uc?export=download&confirm=$(wget --quiet --save-cookies /tmp/cookies.txt --keep-session-cookies --no-check-certificate 'https://docs.google.com/uc?export=download&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx' -O- | sed -rn 's/.*confirm=([0-9A-Za-z_]+).*/\1\n/p')&id=1ymFRBvW1wKIHrdi-v5wyzJLKu81ZxFOx" \
  -O machines.tar && rm -rf /tmp/cookies.txt
sudo tar -xvf machines.tar


mkdir ~/go
mkdir ~/go/src
mkdir ~/go/src/github.com/pfnet-research
cd ~/go/src/github.com/pfnet-research
git clone https://github.com/lenhattan86/k8s-cluster-simulator