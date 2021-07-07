#!/bin/sh

set -e

executable="go-oauth2-server"
cmd="$@"

exec $executable $cmd
