#!/usr/bin/env bash

curl -s "https://gitlab-ci-local-ppa.firecow.dk/pubkey.gpg" | sudo apt-key add -

echo "deb https://gitlab-ci-local-ppa.firecow.dk ./" | sudo tee /etc/apt/sources.list.d/gitlab-ci-local.list

sudo apt-get -yqq update
sudo apt-get -yqq install gitlab-ci-local
