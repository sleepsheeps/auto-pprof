#/bin/bash
sys=$1
if [ -z $sys ]; then
  echo "system is nil"
else
  echo "start build chatd with $sys"
fi


GOOS=$sys go build -mod=mod -o auto-pprof .