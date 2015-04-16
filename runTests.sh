#!/bin/bash

args=${@:-"--short"}
if [ args == "--" ]; then
	args = ""
fi

go test ./... $args
