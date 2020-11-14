#!/usr/bin/env bash

set -e

if [ "$(id -u)" -eq "0" ]; then
  echo "build.sh cannot be run as root"
  exit 1
fi

TIMEOUT=300 # seconds

TEXINPUTS="/latex/texaas:$(kpsepath tex)"
export TEXINPUTS
export TEXMFVAR="/latex/.cache"

USER=$(whoami)
export USER # to avoid latexmk warnings

find /latex/texaas -name "*latexmkrc" -delete

cd "/latex/texaas/$1" || exit 1
rm -f output.pdf output.log

timeout -s KILL "$TIMEOUT" latexmk -r /tools/latexmkrc.pl -pdf "$2" >latexmk.log 2>&1
