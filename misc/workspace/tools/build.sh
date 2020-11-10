#!/usr/bin/env bash

TEXINPUTS="/latex/texaas:$(kpsepath tex)"
export TEXINPUTS
export TEXMFVAR="/latex/.cache"

cd "/latex/texaas/$1" || exit 1
latexmk -r /tools/latexmkrc.pl -pdf "$2"
