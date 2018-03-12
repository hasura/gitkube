#!/usr/bin/env bash


# Teardown resources created by test.bash
USAGE="./$(basename "$0") [optional-directory]"

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

GITKUBE_NAMESPACE_STRING="namespace: kube-system"
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

# read output variables written by test.bash
read_output_vars() {
    GITKUBE_NAMESPACE=$(cat "$OUTPUT_DIR/gitkube-namespace")
    TEST_NAMESPACE=$(cat "$OUTPUT_DIR/test-namespace")
    REMOTE_NAME=$(cat "$OUTPUT_DIR/remote-name")
    TEMP_REPO_DIR=$(cat "$OUTPUT_DIR/temp-repo-dir")
}

# teardown all resources
teardown() {
    log "deleting gitkube resources from $GITKUBE_NAMESPACE"
    cat "$WORKDIR/manifests/gitkube-setup.yaml" | sed -e "s/$GITKUBE_NAMESPACE_STRING/namespace: $GITKUBE_NAMESPACE/" | kctl delete -f -
    kctl delete namespace $GITKUBE_NAMESPACE

    log "deleting test namespace $TEST_NAMESPACE"
    kctl delete namespace $TEST_NAMESPACE

    log "removing temp git repo"
    rm -rf "$TEMP_REPO_DIR"
}

# set kubeconfig and teardown
set_kubeconfig && read_output_vars && teardown
