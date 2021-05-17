#!/usr/bin/env bash

set -eo pipefail

# where am i?
me="$0"
me_home=$(dirname "$0")
me_home=$(cd "$me_home" && pwd)

# deps
DOCKERIZE=dockerize
COMPOSE=docker-compose

# parse arguments
args=$(getopt dcv $*)
set -- $args
for i; do
  case "$i"
  in
    -d)
      debug="true";
      shift;;
    -c)
      other_flags="$other_flags -cover";
      shift;;
    -v)
      other_flags="$other_flags -v";
      shift;;
    --)
      shift; break;;
  esac
done

$COMPOSE -f "$me_home/test/services.compose" up -d
$DOCKERIZE -wait tcp://localhost:59011/ -timeout 30s
go test$other_flags $*
$COMPOSE -f "$me_home/test/services.compose" down
