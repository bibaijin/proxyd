#!/bin/sh

set -xe

# for data connection
# /proxyd -port 8081 -test -upstreams netserver:8081 -cpuprofile /pprof/data.prof -memprofile /pprof/data.mprof -blockprofile /pprof/data.bprof -blockprofilerate 1 &
/proxyd -port 8081 -test -upstreams netserver:8081 &

# for control connection
exec /proxyd -port 8080 -test -upstreams netserver:12865
