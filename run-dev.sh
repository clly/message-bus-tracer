#!/usr/bin/bash

export HONEYCOMB_DATASET="message-bus-tracer"

go build .
./message-bus-tracer

