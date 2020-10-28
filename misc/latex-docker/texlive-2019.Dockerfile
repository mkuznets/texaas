FROM ghcr.io/mkuznets/ubuntu:20.04-2020.09.20

LABEL maintainer="Max Kuznetsov <maks.kuznetsov@gmail.com>"

RUN \
    apt-get update --quiet && \
    \
    PACKAGES=$(apt-cache depends texlive-full | grep Depends | cut -d ':' -f2 | grep -E -v -- '-doc$' | tr '\n' ' ') \
    && \
    apt-get install -f --yes --quiet --no-upgrade --no-install-recommends \
        $PACKAGES \
        biber \
        gnuplot \
    && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
	mkdir /latex

WORKDIR /latex
