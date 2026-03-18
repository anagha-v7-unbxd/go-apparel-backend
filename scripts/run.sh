#!/bin/bash

service="app"

if [[ -z "${SERVICE}" ]]; then 
    echo "Service Environment Variable: \$SERVICE is required"
    exit 1
fi

service="$(echo $SERVICE | tr '[:upper:]' '[:lower:]')"
service_upper="$(echo $SERVICE | tr '[:lower:]' '[:upper:]')"


echo "-----------------------------------------------"
echo "--       $SERVICE Startup Script"
echo "-----------------------------------------------"
echo "Hosts File:"
echo "-----------------------------------------------"
cat /etc/hosts
echo "-----------------------------------------------"
echo "OS-Releases File:"
echo "-----------------------------------------------"
cat /etc/os-release
echo "-----------------------------------------------"
echo "Env details with '$service_upper' prefix."
for en in $(env |grep '^'$service_upper''); do
    echo $en
done
echo "-----------------------------------------------"
echo "> Listing all files in $(pwd)"
ls -ltrh "$(pwd)"
echo "-----------------------------------------------"
echo "setting \$PATH variable = $(pwd)/../bin:$(pwd)/bin"
export PATH="$(pwd)/../bin:$(pwd)/bin:$PATH"
echo "-----------------------------------------------"
echo "> checking binary file"
stat "$(which $service.bin)"
echo "-----------------------------------------------"

if ! command -v "$service.bin" > /dev/null; then
    echo "service: $service.bin not found in \$PATH."
    echo "--"
    echo "PATH variable defined as:"
    echo "--"
    echo "$PATH"
    echo "--"
    exit 1
fi

echo "executing: $service.bin start"
echo "----------------------------------------------------------------"
eval "$service.bin start"
echo "----------------------------------------------------------------"
echo "$service.bin terminated with status code: $?"
