#!/usr/bin/env sh
set -e

# find the docker socket owner group id
DOCKER_SOCK_OWNER_GROUP_ID=$(stat -c '%g' /var/run/docker.sock)
# check the container's groups to see if it has a group with the same id
DOCKER_SOCK_OWNER_GROUP=$(getent group "$DOCKER_SOCK_OWNER_GROUP_ID" | cut -d: -f1)
if [ -z "${DOCKER_SOCK_OWNER_GROUP}" ]; then
    # there is no group in the container with the given group id
    # set owner group as 'docker'
    DOCKER_SOCK_OWNER_GROUP="docker"
    # create a new group with the same group id
    groupadd -g "$DOCKER_SOCK_OWNER_GROUP_ID" "$DOCKER_SOCK_OWNER_GROUP"
fi

if [ -f /sshd-conf/remotes.json ]; then
    GIT_REMOTES_CONF="$(cat /sshd-conf/remotes.json)"
fi
echo $GIT_REMOTES_CONF

if [ "$GIT_REMOTES_CONF" != "null" ]; then

    echo "Setting up git remotes"
    echo $GIT_REMOTES_CONF | jq 'keys' | jq -c -r '.[]' | while read repo; do

        HOME_DIR=/home/$repo

        echo "Configuring : $repo"
        adduser --disabled-login --gecos "$repo" --shell /usr/bin/git-shell "$repo"
        mkdir -p $HOME_DIR/git-shell-commands
        cp /sshd-lib/no-interactive-login.sh $HOME_DIR/git-shell-commands/no-interactive-login

        chown -R $repo:$repo $HOME_DIR/git-shell-commands
        chmod +x $HOME_DIR/git-shell-commands/no-interactive-login

        usermod -aG "$DOCKER_SOCK_OWNER_GROUP" "$repo"

        # Create the .ssh directory if it does not exist
        mkdir -p $HOME_DIR/.ssh

        # Set correct permissions for ssh files
        chmod 700 $HOME_DIR/.ssh

        # Copy the authorized-keys from sshd-conf
        #TODO: get remotes' authorized keys
        AUTHORIZED_KEYS="$(echo $GIT_REMOTES_CONF | jq -r --arg r $repo '.[$r]'.'"authorized-keys"')"
        echo "$AUTHORIZED_KEYS" > $HOME_DIR/.ssh/authorized_keys

        chmod 644 $HOME_DIR/.ssh/authorized_keys
        chown -R $repo:$repo $HOME_DIR/.ssh

        # Generate kubernetes configuration
        mkdir $HOME_DIR/.kube
        cat << EOF > $HOME_DIR/.kube/config
apiVersion: v1
kind: Config
current-context: local-k8s
clusters:
- cluster:
    server: https://kubernetes.default
    certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
  name: k8s-cluster
contexts:
- context:
    cluster: k8s-cluster
    namespace: default
    user: serviceaccount
  name: local-k8s
users:
- name: serviceaccount
  user:
    token: $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
EOF
        # .kube should be owned by hasura
        chown -R $repo:$repo $HOME_DIR/.kube

        REGISTRY_CONF="$(echo $GIT_REMOTES_CONF | jq -r --arg r $repo '.[$r]'.'"registry"')"
        if [ "$REGISTRY_CONF" != "null" ]; then
            echo "registry information is being processed"
            export REGISTRY_PREFIX=$(echo $REGISTRY_CONF | jq -r '.prefix')

            DOCKERCFG=$(echo $REGISTRY_CONF | jq -r '.dockercfg')
            if [ "$DOCKERCFG" != "null" ]; then
                echo "Found secret type .dockercfg"
                echo "writing dockercfg to $HOME_DIR/.dockercfg"
                echo $DOCKERCFG > $HOME_DIR/.dockercfg
                chown -R $repo:$repo $HOME_DIR/.dockercfg
            fi

            DOCKERCONFIGJSON=$(echo $REGISTRY_CONF | jq -r '.dockerconfigjson')
            if [ "$DOCKERCONFIGJSON" != "null" ]; then
                echo "Found secret type .dockerconfigjson"
                echo "writing dockerconfigjson to $HOME_DIR/.docker/config.json"
                mkdir -p $HOME_DIR/.docker
                echo $DOCKERCONFIGJSON > $HOME_DIR/.docker/config.json
                chown -R $repo:$repo $HOME_DIR/.docker/config.json
            fi

        fi

        export MANIFEST_OPTS=$(echo $GIT_REMOTES_CONF | jq -c --arg r $repo '.[$r].manifests')
        export REPO_OPTS=$(echo $GIT_REMOTES_CONF | jq -c --arg r $repo '.[$r].deployments')
        REPO_LOC=$HOME_DIR/git/$repo

        # Create the directory
        mkdir -p $REPO_LOC
        # Initialise bare bones git repo
        git init --bare $REPO_LOC

        # Render the pre-receive script with correct values inside the $repo
        mo /sshd-lib/pre_receive.sh > $REPO_LOC/hooks/pre-receive

        # Set appropriate permissions
        chmod a+x $REPO_LOC/hooks/pre-receive
        chown -R $repo:$repo $REPO_LOC
    done
fi

# Now transfer control to ssh daemon
exec /usr/sbin/sshd -D -e -f /sshd-lib/sshd_config
