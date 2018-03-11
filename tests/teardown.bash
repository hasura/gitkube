#!/usr/bin/env bash

# Teardown resources created by test.bash
USAGE="./$(basename "$0") [optional-directory]"

# Some helpful functions
yell() { echo "FAILED> $*" >&2; }
die() { yell "$*"; exit 1; }
try() { "$@" || die "failed executing: $*"; }
log() { echo "--> $*"; }

# Set WORDKDIR
if [ $# -eq 0 ]; then
    # Get current directory (the code to test reside here)
    WORKDIR=$(pwd)
elif [ $# -eq 1 ]; then
    WORKDIR=$1
else
    die "invalid number of arguments. usage: ${USAGE}"
fi

GITKUBED_NAMESPACE_STRING="namespace: kube-system"
TEST_NAMESPACE_STRING="namespace: default"

# set kubernetes config file
set_kubeconfig() {
    if [ -n "$KUBECONFIG" ]; then
        KCONFIG=${KUBECONFIG}
    else
        KCONFIG=$HOME/.kube/config
    fi
    log "current kubectl context: $(kctl config current-context)"
    log "derived from: $KCONFIG"
}

# kubectl helper function
kctl() {
    kubectl --kubeconfig "$KCONFIG" "$@"
}

# read namespaces writted by test.bash
read_namespaces() {
    GITKUBED_NAMESPACE=$(cat "gitkubed.namespace")
    TEST_NAMESPACE=$(cat "test.namespace")
}

# teardown all resources
teardown() {
    log "deleting gitkubed resources from $GITKUBED_NAMESPACE"
    cat "$WORKDIR/manifests/gitkube-setup.yaml" | sed -e "s/$GITKUBED_NAMESPACE_STRING/namespace: $GITKUBED_NAMESPACE/" | kctl delete -f -
    kctl delete namespace $GITKUBED_NAMESPACE

    log "deleting example deployment from $TEST_NAMESPACE"
    cat "$WORKDIR/example/www/k8s.yaml" | sed -e "s/$TEST_NAMESPACE_STRING/namespace: $TEST_NAMESPACE/" | kctl delete -f -
    kctl delete namespace $TEST_NAMESPACE

    log "removing git remote"
    git remote remove $(cat "remote.name")
}

# set kubeconfig and teardown
set_kubeconfig && read_namespaces && teardown
