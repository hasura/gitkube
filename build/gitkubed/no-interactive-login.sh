#!/bin/sh
echo "Hi, you've successfully authenticated, but there is no shell access."
echo
echo "To access a k8s service locally: "
echo
echo "    $ ssh -p <PORT> -N -L 6432:postgres.hasura:5432 hasura@<project-name>.hasura-app.io"
echo
exit 128
