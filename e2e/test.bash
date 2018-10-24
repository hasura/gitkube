#!/usr/bin/env bash
# Basic e2e test for gitkube workflow

set -Ee

# Some helpful functions
yell() { echo "FAILED> $*" >&2; }
die() { yell "$*"; exit 1; }
try() { "$@" || die "failed executing: $*"; }
log() { echo "--> $*"; }

# set WORDKDIR
# run from project root
WORKDIR=$(dirname ${BASH_SOURCE})/..

# set temporary output directory
OUTPUT_DIR="$WORKDIR/e2e/_output"
# create the directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

GITKUBE_NAMESPACE_STRING="namespace: kube-system"
TEST_NAMESPACE_STRING="namespace: default"

# change to HTTPS url when repo is made public
# GITKUBE_EXAMPLE_REPO="https://github.com/hasura/gitkube-example"
GITKUBE_EXAMPLE_REPO="git@github.com:hasura/gitkube-example.git"


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

# create a namespace to install gitkube
create_gitkube_namespace() {
    GITKUBE_NAMESPACE="gitkube-$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)"
    log "creating gitkube namespace $GITKUBE_NAMESPACE"
    kctl create namespace "$GITKUBE_NAMESPACE"
    echo "$GITKUBE_NAMESPACE" > "$OUTPUT_DIR/gitkube-namespace"
}

# install all gitkube components
install_gitkube() {
    create_gitkube_namespace
    log "creating gikubed and dependent resources"
    cat "$WORKDIR/manifests/gitkube-setup.yaml" | sed -e "s/$GITKUBE_NAMESPACE_STRING/namespace: $GITKUBE_NAMESPACE/" | kctl create -f -
    log "waiting for gitkube to start running"
    kctl --namespace "$GITKUBE_NAMESPACE" rollout status deployment/gitkubed
    kctl --namespace "$GITKUBE_NAMESPACE" rollout status deployment/gitkube-controller
    kctl --namespace "$GITKUBE_NAMESPACE" expose deployment gitkubed --type=LoadBalancer --name=gitkubed

}

create_test_namespace() {
    TEST_NAMESPACE="gitkube-test-$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)"
    log "creating test namespace $TEST_NAMESPACE"
    kctl create namespace "$TEST_NAMESPACE"
    echo "$TEST_NAMESPACE" > "$OUTPUT_DIR/test-namespace"
}

clone_test_repo() {
    TEMP_REPO_DIR="/tmp/gitkube-test-$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)"
    mkdir -p "$TEMP_REPO_DIR"
    echo "$TEMP_REPO_DIR" > "$OUTPUT_DIR/temp-repo-dir"
    git clone "$GITKUBE_EXAMPLE_REPO" "$TEMP_REPO_DIR"
}

# create test deployment
create_test_manifests() {
    create_test_namespace
    clone_test_repo
    log "creating test deployment"
    sed -i -e "s/$TEST_NAMESPACE_STRING/namespace: $TEST_NAMESPACE/" "$TEMP_REPO_DIR/mono-repo/manifests/"*
    git -C "$TEMP_REPO_DIR" commit -am "change to test namespace"
}

# add ssh key to remote config
add_ssh_key() {
    log "adding ssh key to remote file"
    try cat ~/.ssh/id_rsa.pub | awk '$0="  - "$0' >> "$TEMP_REPO_DIR/mono-repo/remote.yaml"
}

# create remote object on the cluster
create_remote() {
    add_ssh_key
    log "creating gitkube remote object"
    cat "$TEMP_REPO_DIR/mono-repo/remote.yaml" | sed -e "s/$TEST_NAMESPACE_STRING/namespace: $TEST_NAMESPACE/" | kctl create -f -
}

# get remote url from the object
get_remote_url() {
    REMOTE_URL=$(kctl --namespace "$TEST_NAMESPACE" get remote example -o json | jq -r '.status.remoteUrl')
    echo "remote url: $REMOTE_URL"
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
    echo "$REMOTE_NAME" > "$OUTPUT_DIR/remote-name"
    try git -C "$TEMP_REPO_DIR" remote add $REMOTE_NAME "$REMOTE_URL"
}

# git push to the remote
git_push() {
    try git -C "$TEMP_REPO_DIR" push --no-verify $REMOTE_NAME master
}

# run basic test
run_basic_test() {
    ensure_dependencies
        set_kubeconfig
        install_gitkube
        create_test_manifests
        create_remote
        setup_local_remote
        git_push
# TODO: test 'edit-and-push' flow
}

# execute the test
run_basic_test

echo
echo "TEST PASSED"
echo
echo "run teardown"
