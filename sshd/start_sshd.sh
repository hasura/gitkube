#!/usr/bin/env sh
set -e

export HOST_GROUP_ID=$(cat /hasura-data/group | grep '^docker' | cut -d: -f3)
GROUP_WITH_HOST_GROUP_ID=$(getent group $HOST_GROUP_ID | cut -d: -f1)
if [ -z "${GROUP_WITH_HOST_GROUP_ID}" ]; then
    # Find the group id from the host and use it to create docker group
    groupadd -g $HOST_GROUP_ID docker
    GROUP_WITH_HOST_GROUP_ID="docker"
fi

GIT_REMOTES_CONF="$(cat /sshd-conf/remotes.json)"
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

        usermod -aG $GROUP_WITH_HOST_GROUP_ID $repo

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
            echo "writing dockercfg to $HOME_DIR/.dockercfg"
            echo $DOCKERCFG > $HOME_DIR/.dockercfg
            chown -R $repo:$repo $HOME_DIR/.dockercfg
        fi

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
