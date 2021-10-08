FROM ghcr.io/estuary/flow:dev

# Switch to the root user to install packages and setup sudo access
USER root
RUN apt-get update -y \
    && apt-get install --no-install-recommends -y \
      postgresql-client \
      sqlite3 \
      jq \
      sudo \
    && rm -rf /var/lib/apt/lists/*

# Allow users to use sudo inside the dev container. This allows people to install their own
# software in the container without modifying this Dockerfile.
RUN echo flow ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/flow \
    && chmod 0440 /etc/sudoers.d/flow

# Generate Flow catalog JSON-Schema for integration with VSCode.
RUN flowctl json-schema > /usr/local/etc/flow.schema.json

# Enable `flow` user to access /var/run/docker.sock
# Adapted from: https://github.com/microsoft/vscode-dev-containers/tree/main/containers/docker-from-docker#adding-the-user-to-a-docker-group
RUN echo "#!/bin/sh\n\
    sudoIf() { if [ \"\$(id -u)\" -ne 0 ]; then sudo \"\$@\"; else \"\$@\"; fi }\n\
    SOCKET_GID=\$(stat -c '%g' /var/run/docker.sock) \n\
    if [ \"${SOCKET_GID}\" != '0' ]; then\n\
        if [ \"\$(cat /etc/group | grep :\${SOCKET_GID}:)\" = '' ]; then sudoIf groupadd --gid \${SOCKET_GID} docker-host; fi \n\
        if [ \"\$(id flow | grep -E \"groups=.*(=|,)\${SOCKET_GID}\(\")\" = '' ]; then sudoIf usermod -aG \${SOCKET_GID} flow; fi\n\
    fi\n\
    exec \"\$@\"" > /usr/local/share/docker-init.sh \
    && chmod +x /usr/local/share/docker-init.sh

# VS Code overrides ENTRYPOINT and CMD when executing `docker run` by default.
# Setting the ENTRYPOINT to docker-init.sh will configure non-root access to
# the Docker socket if "overrideCommand": false is set in devcontainer.json.
# The script will also execute CMD if you need to alter startup behaviors.
ENTRYPOINT [ "/usr/local/share/docker-init.sh" ]
CMD [ "sleep", "infinity" ]

USER flow
