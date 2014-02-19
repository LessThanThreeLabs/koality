#!/bin/sh
#
# Koality service
#
# chkconfig: - 95 05
# description: Never break a build again
#

### BEGIN INIT INFO
# Provides:          koality
# Required-Start:    postgresql $remote_fs
# Required-Stop:     $remote_fs
# Should-Stop:       postgresql
# Default-Start: 2 3 4 5
# Default-Stop: 0 1 6
# Description:       Koality
# Short-Description: Never break a build again
### END INIT INFO

PIDFILE=/var/run/koality.pid

start_koality () {
	start-stop-daemon --start --pidfile $PIDFILE --startas /etc/koality/current/code/back/bin/koalityRunner --background --make-pidfile
}

stop_koality () {
	 start-stop-daemon --stop --pidfile $PIDFILE --retry 10
}

koality_status () {
	start-stop-daemon --status --pidfile $PIDFILE
}

restart_koality () {
	stop_koality
	start_koality
}

case "$1" in
	start)
		start_koality
		;;
	stop)
		stop_koality
		;;
	restart)
		restart_koality
		;;
	status)
		koality_status
		;;
	*)
		echo "Please use start, stop, restart, or status as first argument"
		;;
esac
