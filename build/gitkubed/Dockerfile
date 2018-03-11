FROM debian:stretch-20171009
MAINTAINER vamshi@34cross.in

# Install openssh server
RUN apt-get update \
 && apt-get install -y upx-ucl binutils curl openssh-server git jq \
 && curl -o /tmp/docker-1.12 'https://get.docker.com/builds/Linux/x86_64/docker-1.12.6.tgz' \
 && tar -xf /tmp/docker-1.12 -C /tmp \
 && mv /tmp/docker/docker /bin/docker \
 && rm -rf /tmp/docker-1.12 /tmp/docker \
 && strip --strip-unneeded /bin/docker \
 && chmod a+x /bin/docker \
 && upx /bin/docker \
 && curl -o /bin/kubectl 'https://storage.googleapis.com/kubernetes-release/release/v1.8.7/bin/linux/amd64/kubectl' \
 && strip --strip-unneeded /bin/kubectl \
 && chmod a+x /bin/kubectl \
 && upx /bin/kubectl \
 && curl -o /usr/bin/mo 'https://raw.githubusercontent.com/tests-always-included/mo/master/mo' \
 && chmod a+x /usr/bin/mo \
 && apt-get purge -y --auto-remove upx-ucl binutils \
 && rm -rf /var/lib/apt/lists/*

# sshd config
RUN mkdir /var/run/sshd

ADD no-interactive-login.sh /sshd-lib/
ADD pre_receive.sh /sshd-lib/
ADD sshd_config /sshd-lib/
ADD start_sshd.sh /sshd-lib/

EXPOSE 22
