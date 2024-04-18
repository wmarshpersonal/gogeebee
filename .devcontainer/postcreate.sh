#!/bin/bash

# build wla-dx
pushd .
cd $(mktemp -d)
git clone https://github.com/vhelin/wla-dx.git
cd wla-dx
git checkout v10.6
cmake -G "Unix Makefiles"
make
sudo make install
popd