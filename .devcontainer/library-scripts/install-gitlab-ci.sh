#!/usr/bin/env bash

curl -s "https://firecow.github.io/gitlab-ci-local/ppa/pubkey.gpg" | sudo apt-key add -
sudo curl -s -o /etc/apt/sources.list.d/gitlab-ci-local.list "https://firecow.github.io/gitlab-ci-local/ppa/gitlab-ci-local.list"
sudo apt-get -yqq update
sudo apt-get -yqq install gitlab-ci-local
