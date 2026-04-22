# -*- mode: shell-script -*-
#
# Copyright 2021-2023 Open Text
# Copyright 2026 Denis Patrakhin (modifications to this Dockerfile)
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
# https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Based on: https://github.com/vertica/vertica-containers/blob/master/one-node-ce/Dockerfile_AlmaLinux

ARG os_version
ARG os_image

FROM ${os_image}:${os_version} AS builder

ENV VERTICA_OPT_DIR="/opt/vertica"
ENV VERTICA_VOLUME_DIR="/data"
ENV VERTICA_DATA_DIR="${VERTICA_VOLUME_DIR}/vertica"
ENV VMART_DIR="${VERTICA_OPT_DIR}/examples/VMart_Schema"

ARG vertica_package
ARG vertica_db_user="dbadmin"
ARG vertica_db_group="verticadba"
ARG vertica_db_name="dockerdb"
ARG dbadmin_gid=1000
ARG dbadmin_id=1000
# Start / end years generated into date dimension table
ARG vmart_start_year=2003
ARG vmart_end_year=2032

ADD ./packages/${vertica_package} /tmp/${vertica_package}
# this is a script which removes unnecessary stuff from the
# container image
ADD ./tools/cleanup.sh /tmp/
ADD ./tools/package-checksum-patcher.py /tmp/
ADD ./tools/init.d.functions /etc/rc.d/init.d/functions

# Install prerequisites and Vertica
    # Update is needed to be confident that we're picking up
    # fixed libraries.  We depend on malware check of container afterwards 
# this creates multiple layers, but the many layers will be taken care of
# when we switch to the next stage of the build
RUN bash -c "if [ \"$(rpm -E %{rhel})\" == '8' ]; then yum install -q -y dnf-plugins-core; yum -q config-manager --set-enabled powertools; yum -q -y install glibc-locale-source; fi"
 # Create DB user and group in OS. Fix their ID, so they match in the following stage
RUN yum -q -y update
RUN yum -y update --security
RUN yum -q -y makecache
RUN yum install -y \
        cronie \
        dialog \
        gcc-c++ \
        glibc-static \
        gdb \
        iproute \
        libstdc++-static \
        make \
        mcelog \
        openssh-server \
        openssh-clients \
        openssh \
        openssl \
        sudo \
        sysstat \
        wget \
        which
RUN yum localinstall -q -y /tmp/${vertica_package}
 # Cleanup to optimize the size of final image
RUN rm -rf ${VERTICA_OPT_DIR}/lib64
RUN bash -c "if [ \"$(rpm -E %{rhel})\" == '8' ]; then localedef -i en_US -f UTF-8 en_US.UTF-8; fi"
# Create DB user in OS and create python egg cache dir under this user
# Run install_vertica script to prepare environment
RUN mkdir -p ${VERTICA_DATA_DIR} \
     && /usr/sbin/groupadd -r -g $dbadmin_gid $vertica_db_group \
     # While it is true that install_vertica creates $vertica_db_user
     # and $vertica_db_group we do it by hand here in order to be able
     # to specify the UID of the $vertica_db_user  
     && /usr/sbin/useradd -r \
                          -m \
                          -s /bin/bash \
                          --uid $dbadmin_id \
                          -g $vertica_db_group \
                          $vertica_db_user \
     && chown -R $vertica_db_user:$vertica_db_group ${VERTICA_DATA_DIR}

# Compile VMART bins and generate date dimension corresponding to
# start/end years coming from ARGs  
RUN set -x \
 && cd ${VMART_DIR} \
 && make vmart_gen \
 && g++ vmart_genTime.c -o vmart_genTime \
 && chmod +x vmart_genTime \
 && ./vmart_genTime ${vmart_start_year} ${vmart_end_year} > Time_custom.txt

RUN chown -R $vertica_db_user:$vertica_db_group ${VERTICA_OPT_DIR} \
    && chown -R $vertica_db_user:$vertica_db_group /home/$vertica_db_user

# docker daemons versions < 19.03 don't preserve ownership on COPY --from
# but doing the chown -R in the second half is costly in terms of
# image size, so we compromise by making everything group writable in
# /opt/vertica:
RUN chmod -R g+w ${VERTICA_OPT_DIR}

RUN sh /tmp/cleanup.sh

############################################################################
FROM ${os_image}:${os_version}

# Args used only in FROM lines can be declared before the first FROM; repeat here
# for LABEL and any later instructions in this stage.
ARG os_version

ARG vertica_db_user="dbadmin"
ARG vertica_db_group="verticadba"
ARG vertica_db_name="dockerdb"

ARG dbadmin_gid=1000
ARG dbadmin_id=1000
# Same RPM as builder: install_vertica --rpm for peers; local rpmdb entry via rpm --justdb (see RUN below).
ARG vertica_package

ENV VERTICA_OPT_DIR="/opt/vertica"
ENV VERTICA_DB_USER="${vertica_db_user}"
ENV VERTICA_DB_NAME="$vertica_db_name"
ENV VERTICA_OPT_BIN="${VERTICA_OPT_DIR}/bin"
ENV VERTICA_HOME_DIR="/home/$vertica_db_user"
ENV VERTICA_VOLUME_DIR="/data"
ENV VERTICA_DATA_DIR="${VERTICA_VOLUME_DIR}/vertica"
ENV PYTHON_EGG_CACHE="/tmp/.python-eggs"
# https://www.vertica.com/docs/11.0.x/HTML/Content/Authoring/GettingStartedGuide/IntroducingVMart/IntroducingVMart.htm?zoom_highlight=vmart
ENV VMART_DIR="${VERTICA_OPT_DIR}/examples/VMart_Schema"
ENV VMART_CONFIRM_LOAD_SCHEMA="public"
ENV VMART_CONFIRM_LOAD_TABLE="vmart_load_success"
ENV VMART_ETL_SCRIPT="01_load_vmart_schema.sh"
ENV VMART_ETL_SQL="02_vmart_etl.sql"
ENV ENTRYPOINT_SCRIPT="docker-entrypoint.sh"
ENV ENTRYPOINT_SCRIPT_PATH="${VERTICA_HOME_DIR}/${ENTRYPOINT_SCRIPT}"
# install_vertica needs --rpm to push to other nodes; local host must also show vertica in rpmdb.
ENV VERTICA_INSTALL_RPM_PATH="${VERTICA_OPT_DIR}/install-rpm/${vertica_package}"

COPY --from=builder $VERTICA_OPT_DIR $VERTICA_OPT_DIR
COPY --from=builder $VERTICA_DATA_DIR $VERTICA_DATA_DIR
COPY --from=builder /tmp/${vertica_package} ${VERTICA_OPT_DIR}/install-rpm/${vertica_package}

ADD ./tools/package-checksum-patcher.py /tmp/
ADD ./tools/init.d.functions /etc/rc.d/init.d/functions
# SSH client: skip host-key prompts for default three-node compose (root + dbadmin; not ~/.ssh/config).
ADD ./env_setup/ssh.conf /etc/ssh/ssh_config.d/90-vertica-docker-cluster.conf

# Recreate DB user in OS and create python egg cache dir under this user
# in this stage of the build, we want to reduce the number of layers
RUN set -x \
    # COPY doesn't preserve the mode of the top-level
    && chmod g+w ${VERTICA_OPT_DIR} \
    && /usr/sbin/groupadd -r -g $dbadmin_gid $vertica_db_group \
    && /usr/sbin/useradd -r \
                          -m \
                          -s /bin/bash \
                          --uid $dbadmin_id \
                          # add dbadmin to root to work around COPY
                          # not preserving owner and group prior to
                          # docker version 19.03
                          -G root \
                          -g $vertica_db_group \
                          $vertica_db_user \
 && su - $vertica_db_user -c "mkdir -p $PYTHON_EGG_CACHE" \
 && su - $vertica_db_user -c "mkdir -p /home/$vertica_db_user/.ssh && chmod 700 /home/$vertica_db_user/.ssh" \
# Install prerequisites, create Vertica data dir, set privileges, document
# fix timezone issue and generate locales in case of centOS 8 
 && yum -q -y update \
 && yum -y update --security \
 && yum -q -y makecache \
 && bash -c "if [ \"$(rpm -E %{rhel})\" == '8' ]; then yum install -q -y dnf-plugins-core; yum -q config-manager --set-enabled powertools; yum -q -y install glibc-locale-source; localedef -i en_US -f UTF-8 en_US.UTF-8; fi" \
     && yum install -y \
        cronie \
        dialog \
        gdb \
        iproute \
        libstdc++-static \
        mcelog \
        openssh-server \
        openssh-clients \
        openssh \
        openssl \
        sudo \
        sysstat \
        wget \
        which \
        procps-ng \
    # Allow passwordless sudo access from dbadmin
    && echo "$vertica_db_user ALL=(ALL) NOPASSWD: ALL" | tee -a /etc/sudoers \
    && /usr/bin/ssh-keygen -A \
    # docker daemons versions < 19.03 don't preserve ownership on COPY --from
    # so we have to chown these things again
    # Pre-create catalog + data dirs for ${vertica_db_name} under VERTICA_DATA_DIR
    # so the named /data volume (declared below) seeds them on first compose up,
    # removing the need for a separate create-db-dirs step before admintools.
    && mkdir -p ${VERTICA_DATA_DIR}/${vertica_db_name}_catalog \
                ${VERTICA_DATA_DIR}/${vertica_db_name}_data \
    && chown -R $vertica_db_user:$vertica_db_group ${VERTICA_DATA_DIR} \
    && chmod 0755 ${VERTICA_DATA_DIR}/${vertica_db_name}_catalog \
                  ${VERTICA_DATA_DIR}/${vertica_db_name}_data \
        && chown -R $vertica_db_user:$vertica_db_group /home/$vertica_db_user \
    # but this chown is expensive in image-size, so we use aother
    # trick in the first half of the build 
    # && chown -R $vertica_db_user:$vertica_db_group ${VERTICA_OPT_DIR} \
    # Link OS timezones, check entrypoint script
    # && rm -rf ${VERTICA_OPT_DIR}/share/timezone \
    # && ln -sT /usr/share/zoneinfo ${VERTICA_OPT_DIR}/share/timezone \
    # Files already from builder COPY; only update rpmdb so install_vertica local check passes (avoids "upgrade locally" + script bug).
    && rpm -ivh --nodeps --nosignature --noscripts --justdb "/opt/vertica/install-rpm/${vertica_package}" \
        && yum clean all \
        && rm -rf /var/cache/yum

ADD ./env_setup/vertica_env.sh /etc/profile.d/
ADD ./vmart/${VMART_ETL_SQL} ./vmart/${VMART_ETL_SCRIPT} ${VMART_DIR}/
RUN chmod 0755 ${VMART_DIR}/${VMART_ETL_SCRIPT} \
 && chown ${vertica_db_user}:${vertica_db_group} \
          ${VMART_DIR}/${VMART_ETL_SQL} \
          ${VMART_DIR}/${VMART_ETL_SCRIPT}

ADD ./${ENTRYPOINT_SCRIPT} ${ENTRYPOINT_SCRIPT_PATH}
RUN chmod +x ${ENTRYPOINT_SCRIPT_PATH}

ADD ./env_setup/.vsqlrc /home/${vertica_db_user}/.vsqlrc
RUN chown ${vertica_db_user}:${vertica_db_group} /home/${vertica_db_user}/.vsqlrc

# SSH (local dev only — do not push images with real keys).
# Parity with stats-ansible vertica_create_cluster/tasks/create_cluster.yml:
#   - root: /root/.ssh (0700), id_rsa (0600) = cluster key for install_vertica; authorized_keys = extra client keys.
#     Ansible adds deploy/odin/vertica pubkeys; here one .pub seeds authorized_keys (add more lines manually if needed).
#   - dbadmin: no files in ~/.ssh (only empty dir 0700 from RUN above). install_vertica creates keys; pre-baking
#     id_rsa/authorized_keys here conflicts with that step and can cause "SSH key generation or distribution failed".
# From host: root SSH client keys only (.ssh/id_rsa + .ssh/id_rsa.pub). sshd host keys: ssh-keygen -A in RUN above.
COPY .ssh/id_rsa .ssh/id_rsa.pub /tmp/vertica-baked-ssh/
RUN mkdir -p /root/.ssh \
    && chmod 700 /root/.ssh \
    && install -m 600 -o root -g root /tmp/vertica-baked-ssh/id_rsa /root/.ssh/id_rsa \
    && install -m 644 -o root -g root /tmp/vertica-baked-ssh/id_rsa.pub /root/.ssh/id_rsa.pub \
    && install -m 600 -o root -g root /tmp/vertica-baked-ssh/id_rsa.pub /root/.ssh/authorized_keys \
    && rm -rf /tmp/vertica-baked-ssh

WORKDIR /root

VOLUME ${VERTICA_VOLUME_DIR}

ENTRYPOINT $ENTRYPOINT_SCRIPT_PATH

# ssh port
EXPOSE 22
# vertica port
EXPOSE 5433
# vertica agent port
EXPOSE 5444

LABEL image_name="vertica_node"
LABEL os_family="almalinux"
LABEL os_version="$os_version"
LABEL maintainer="patraden"
LABEL org.opencontainers.image.licenses="Apache-2.0"
