#!/usr/bin/env bash

set -Ee

# Basic e2e test for gitkube workflow
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

# check if the system has the given command
system_has() {
    type "$1" > /dev/null 2>&1
}

# ensure that kubectl, git and jq are present
ensure_dependencies() {
    if ! system_has "kubectl"; then
        die "kubectl not found"
    fi
    if ! system_has "git"; then
        die "git not found"
    fi
    if ! system_has "jq"; then
        die "jq not found"
    fi
    return 0
}

# set kubernetes config file
# KUBECONFIG env var is taken
set_kubeconfig() {
    if [ -n "$KUBECONFIG" ]; then
        KCONFIG=${KUBECONFIG}
    else
        KCONFIG=$HOME/.kube/config
    fi
    log "current kubectl context: $(kctl config current-context)"
    log "derived from: $KCONFIG"
}

# kubectl command helper
kctl() {
    try kubectl --kubeconfig "$KCONFIG" "$@"
}

# create a namespace to install gitkubed
create_gitkubed_namespace() {
    GITKUBED_NAMESPACE="gitkubed-$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)"
    log "creating gitkubed namespace $GITKUBED_NAMESPACE"
    kctl create namespace "$GITKUBED_NAMESPACE"
    echo "$GITKUBED_NAMESPACE" > "gitkubed.namespace"
}

# install all gitkube components
install_gitkube() {
    create_gitkubed_namespace
    log "creating gikubed and dependent resources"
    cat "$WORKDIR/manifests/gitkube-setup.yaml" | sed -e "s/$GITKUBED_NAMESPACE_STRING/namespace: $GITKUBED_NAMESPACE/" | kctl create -f -
    log "waiting for gitkubed to start running"
    kctl --namespace "$GITKUBED_NAMESPACE" rollout status deployment/gitkubed
}

create_test_namespace() {
    TEST_NAMESPACE="gitkubed-test-$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)"
    log "creating test namespace $TEST_NAMESPACE"
    kctl create namespace "$TEST_NAMESPACE"
    echo "$TEST_NAMESPACE" > "test.namespace"
}

# create test deployment
create_deployment() {
    create_test_namespace
    log "creating test deployment"
    cat "$WORKDIR/example/www/k8s.yaml" | sed -e "s/$TEST_NAMESPACE_STRING/namespace: $TEST_NAMESPACE/" | kctl create -f -
    log "waiting for test deployment to start running"
    kctl --namespace "$TEST_NAMESPACE" rollout status deployment/www
}

# add ssh key to remote config
add_ssh_key() {
    log "adding ssh key to remote file"
    try cat ~/.ssh/id_rsa.pub | awk '$0="  - "$0' >> "$WORKDIR/example/remote.yaml"
}

# create remote object on the cluster
create_remote() {
    add_ssh_key
    log "creating gitkube remote object"
    cat "$WORKDIR/example/remote.yaml" | sed -e "s/$TEST_NAMESPACE_STRING/namespace: $TEST_NAMESPACE/" | kctl create -f -
}

# get remote url from the object
get_remote_url() {
    REMOTE_URL=$(kctl --namespace "$TEST_NAMESPACE" get remote sampleremote -o json | jq -r '.status.remoteUrl')
    echo $REMOTE_URL
}

# wait for the remote url to be available
MAX_RETRIES=5
RETRY_INTERVAL=10
RETRIES=0
wait_for_remote_url() {
    get_remote_url
    if [ -z "$REMOTE_URL" ] || [ "$REMOTE_URL" == "null" ]; then
        if [ "$RETRIES" == "$MAX_RETRIES" ]; then
            die "could not get remote url"
        else
            RETRIES=$((RETRIES++))
            sleep $RETRY_INTERVAL
            wait_for_remote_url
        fi
    fi
}

# setup remote on the local git repo
setup_local_remote() {
    log "waiting for remote url to be generated"
    wait_for_remote_url
    log "creating git remote"
    REMOTE_NAME="remote-$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)"
    echo "$REMOTE_NAME" > "remote.name"
    try git remote add $REMOTE_NAME "$REMOTE_URL"
}

# git push to the remote
git_push() {
    try git push $REMOTE_NAME master
}

# run basic test
run_basic_test() {
    ensure_dependencies
        set_kubeconfig
        install_gitkube
        create_deployment
        create_remote
        setup_local_remote
        git_push
# TODO: test 'edit-and-push' flow
}

# execute the main function
run_basic_test
