#! /usr/bin/env bash

set -e

go build
sudo cp -rf subgrab /usr/bin/