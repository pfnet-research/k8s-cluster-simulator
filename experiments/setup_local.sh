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
cd /proj/yarnrm-GP0
git clone https://github.com/lenhattan86/google_data_trace