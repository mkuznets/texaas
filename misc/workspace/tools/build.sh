#!/usr/bin/env bash

TEXINPUTS="/latex/texaas:$(kpsepath tex)"
export TEXINPUTS
export TEXMFVAR="/latex/.cache"

find /latex/texaas -name "*latexmkrc" -delete

cd "/latex/texaas/$1" || exit 1
rm -f output.pdf output.log

latexmk -r /tools/latexmkrc.pl -pdf "$2"
