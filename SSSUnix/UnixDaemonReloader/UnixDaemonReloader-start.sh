#!/bin/sh

mkdir -p /var/log/SSS
mkdir -p /var/run/SSS

cd /server/SSS/UnixDaemonReloader

sleep 30 && /server/SSS/UnixDaemonReloader/UnixDaemonReloader.gl -config=/server/SSS/UnixDaemonReloader/UnixDaemonReloader.json -daemon=YES &

