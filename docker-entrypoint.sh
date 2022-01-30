#!/bin/sh
set -e

# If they call register/check/list, let's forward to acme-dns-client directly
if [ "$1" = 'register' ] || [ "$1" = 'check' ] || [ "$1" = 'list' ]; then
    exec acme-dns-client "$@"
fi

# If they provide no arguments, we'll assume crond -f
if [ $# -eq 0 ]; then
    # We can't just do exec crond -f, as signals will get lost and we can't shutdown
    crond -b -L /var/log/cron.log
    tail -f /var/log/cron.log &
    wait $!
    exit 0
fi

# Otherwise run certbot like normal
exec certbot "$@"
