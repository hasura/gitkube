#!/usr/bin/env bash
#
# Author: "FRITZ Thomas" <fritztho@gmail.com> (http://www.fritzthomas.com)
# GitHub: https://gist.github.com/thomasfr/9691385
#
# The MIT License (MIT)
#
# Copyright (c) 2014-2015 FRITZ Thomas
#
# Copyright  2018 The Gitkube authors
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.



# When receiving a new git push, the received branch gets compared to this one.
# If you do not need this, just add a comment
export DEPLOY_ALLOWED_BRANCH="master"

# Repo Name:
export DEPLOY_REPO_NAME=$(basename "$PWD")

REPO_OPTS='{{REPO_OPTS}}'
REGISTRY_PREFIX='{{REGISTRY_PREFIX}}'

export DEPLOYMENTS=$(echo ${REPO_OPTS} | jq -c 'keys' | jq -r '.[]')

# This is the root deploy dir.
export BUILD_ROOT="${HOME}/build/${DEPLOY_REPO_NAME}"

###########################################################################################

# export GIT_DIR="$(cd $(dirname $(dirname $0));pwd)"
# export GIT_WORK_TREE="${BUILD_ROOT}"

echo "Gitkube build system : $(date): Initialising"
echo

# Create the build directory
echo "Creating the build directory"
mkdir -p "${BUILD_ROOT}"

# Loop, because it is possible to push more than one branch at a time. (git push --all)
while read oldrev newrev refname
do

    # export DEPLOY_BRANCH=$(git rev-parse --symbolic --abbrev-ref $refname)
    export DEPLOY_BRANCH=$(expr "$refname" : "refs/heads/\(.*\)")
    export DEPLOY_OLDREV="$oldrev"
    export DEPLOY_NEWREV="$newrev"
    export DEPLOY_REFNAME="$refname"

    if [ ! -z "${DEPLOY_ALLOWED_BRANCH}" ]; then
        if [ "${DEPLOY_ALLOWED_BRANCH}" != "$DEPLOY_BRANCH" ]; then
            echo "Ignoring branch '$DEPLOY_BRANCH' of '${DEPLOY_REPO_NAME}'."
            echo "Deployment(s) will not be updated"
            continue
        fi
    fi

    echo "Checking out '${DEPLOY_BRANCH}:${DEPLOY_NEWREV}' to '${BUILD_ROOT}'"
    git archive $DEPLOY_NEWREV | tar -x -C $BUILD_ROOT

#    export PRE_BUILD_HOOK="${BUILD_ROOT}/.hasura/pre-build"
#
#    if [ -f $PRE_BUILD_HOOK ]; then
#        echo
#        echo "Executing pre-build hook"
#        $PRE_BUILD_HOOK || exit 1
#    fi

    echo ""
    echo "$(echo $DEPLOYMENTS | wc -w) deployment(s) found in this repo"
    echo "Trying to build them..."
    echo ""

    # Now for each deployment in this repo - build the docker images (there can
    # be multiple images in one deployment), and finally update the k8s deployment
    for Q_DEPLOYMENT_NAME in $DEPLOYMENTS; do # qualified deployment name

        # namespace of the deployment
        export K8S_NS=$(echo $Q_DEPLOYMENT_NAME | cut -d '.' -f 1)

        # deployment w/o namespace from qualified deployment name
        export DEPLOYMENT_NAME=$(echo $Q_DEPLOYMENT_NAME | cut -d '.' -f 2)

        echo "Building Docker image for : ${DEPLOYMENT_NAME}"

        DEPL_OPTS=$(echo ${REPO_OPTS} | jq -c --arg d $Q_DEPLOYMENT_NAME '.[$d]')
        DOCKER_IMAGES=$(echo ${DEPL_OPTS} | jq -c 'keys' | jq -r '.[]')

        # Accumulator to gather all the container names and their corresponding
        # images. Used later to update a k8s deployment.
        K8S_DEPL_IMAGE_SET=""

        # there can be multiple images in one deployment - hence loop and build
        # docker images
        for IMAGE_NAME in $DOCKER_IMAGES; do

            DEPL_DOCKER_CONTEXT=$(echo $DEPL_OPTS | jq -c -r --arg i $IMAGE_NAME '.[$i].path')
            DEPL_DOCKER_FILE_PATH=$(echo $DEPL_OPTS | jq -c -r --arg i $IMAGE_NAME '.[$i].dockerfile')
            NO_CACHE=$(echo $DEPL_OPTS | jq -c -r --arg i $IMAGE_NAME '.[$i].noCache')

            # If dockerfile key is not present in the config, assume default
            # Dockerfile in the context path
            if [ "${DEPL_DOCKER_FILE_PATH}" == "null" ]; then
                DEPL_DOCKER_FILE_PATH="${DEPL_DOCKER_CONTEXT}/Dockerfile"
            fi

            DOCKERFILE_PATH="${BUILD_ROOT}/${DEPL_DOCKER_FILE_PATH}"
            DOCKER_BUILD_CONTEXT="${BUILD_ROOT}/${DEPL_DOCKER_CONTEXT}"

            export DEPLOY_IMAGE_NAME="${DEPLOY_REPO_NAME}-${Q_DEPLOYMENT_NAME}-${IMAGE_NAME}"

            if [ ! -f ${DOCKERFILE_PATH} ]; then
                echo "No Dockerfile present. Exiting"
                exit 1
            fi

            echo ""
            export CUR_IMAGE="${DEPLOY_IMAGE_NAME}:${DEPLOY_NEWREV}"

            if [ -n "$REGISTRY_PREFIX" ]; then
                export CUR_IMAGE="${REGISTRY_PREFIX}/${CUR_IMAGE}"
            fi

            NO_CACHE_ARGS=""
            if [ "${NO_CACHE}" = "true" ]; then
                NO_CACHE_ARGS="--no-cache"
            fi
            echo "Building Docker image : ${CUR_IMAGE}"
            docker build $NO_CACHE_ARGS -t "${CUR_IMAGE}" -f "${DOCKERFILE_PATH}" "${DOCKER_BUILD_CONTEXT}" || exit 1
            if [ -n "$REGISTRY_PREFIX" ]; then
                echo "pushing ${CUR_IMAGE} to registry"
                docker push "${CUR_IMAGE}" || exit 1
            fi

            # Keep appending "container-name=docker-image-name" in a string.
            # Used later to update k8s deployment.
            K8S_DEPL_IMAGE_SET="${K8S_DEPL_IMAGE_SET} ${IMAGE_NAME}=${CUR_IMAGE}"
        done

        echo ""
        echo "Updating Kubernetes deployment: $DEPLOYMENT_NAME"
        kubectl --namespace=$K8S_NS set image --record deployment/${DEPLOYMENT_NAME} ${K8S_DEPL_IMAGE_SET} || exit 1

        timeout 60s kubectl --namespace=$K8S_NS rollout status deployment/${DEPLOYMENT_NAME}
        if [ "$?" -ne "0" ]; then
            echo ""
            echo ""
            echo -e "\e[31mUpdating deployment failed: $DEPLOYMENT_NAME \e[0m"
            echo -e "  \e[36m$ Check deployment logs \e[0m"
            echo ""
            echo ""
            exit 1
        fi

        kubectl --namespace=$K8S_NS get deployment ${DEPLOYMENT_NAME} || exit 1
    done

    echo ""
    echo "Removing build directory"
    rm -rf $BUILD_ROOT
done

echo
echo "Gitkube build system : $(date): Finished build"
echo ""
echo ""
exit 0
