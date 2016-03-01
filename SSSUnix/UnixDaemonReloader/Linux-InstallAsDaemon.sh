#!/bin/sh

ln -s /server/SSS/UnixDaemonReloader/UnixDaemonReloader-start.sh /etc/init.d/SSSUDR.sh
update-rc.d -f SSSUDR.sh defaults
