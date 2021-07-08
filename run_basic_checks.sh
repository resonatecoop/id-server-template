#!/usr/bin/env bash
set -e

# get relevant source files
SOURCES=$(find ./ -iname "*.go")
if [ -z "$SOURCES" ]; then
  exit 0
fi

# track which files need to be formated
TOFORMAT=$(gofmt -l $SOURCES)
if [ -z "$TOFORMAT" ]; then
  exit 0
fi

# print which files need to be formatted
echo >&2 "Go files must be formatted with gofmt. Please run:"
for FILE in $TOFORMAT; do
	echo >&2 "  gofmt -w $FILE"
done
exit 23
