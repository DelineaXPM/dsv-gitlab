# Note: You can use any Debian/Ubuntu based image you want.
FROM mcr.microsoft.com/vscode/devcontainers/universal:2-focal


# Enable new "BUILDKIT" mode for Docker CLI
ENV DOCKER_BUILDKIT=1

# Options
ARG INSTALL_ZSH="true"
ARG UPGRADE_PACKAGES="false"
ARG USE_MOBY="true"
ARG USERNAME=codespace
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG ENABLE_NONROOT_DOCKER="true"
ARG DOCKER_VERSION="latest"
ARG GO_VERSION="${GO_VERSION}"
ARG GOPATH="${GOPATH}"
ARG GOROOT="${GO_ROOT}"

# Install needed packages and setup non-root user. Use a separate RUN statement to add your own dependencies.
USER root
ENV DOCKER_BUILDKIT=1
ENV GOPATH=${GOPATH}
ENV PATH="${GOPATH}/bin:${PATH}"
ENV MAGEFILE_ENABLE_COLOR=1
ENV TRUNK_LAUNCHER_QUIET=true

USER ${USERNAME}
RUN mkdir -p /home/${USERNAME}/.minikube \
  && chown ${USERNAME} /home/${USERNAME}/.minikube \
  && /bin/bash -c 'set -euo pipefail && curl https://get.trunk.io -fsSL | bash -s -- -y' \
  && /bin/bash -c 'set -euo pipefail && curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v1.0.0/aqua-installer | bash'
COPY files/.zshrc /home/${USERNAME}/.zshrc
COPY files/first-run-notice.txt /home/${USERNAME}/first-run-notice.txt

VOLUME [ "/var/lib/docker" ]
ENTRYPOINT ["/usr/local/share/docker-init.sh"]
CMD ["sleep", "infinity"]
RUN curl -sfL https://direnv.net/install.sh | bash

USER ${USERNAME}
RUN go install github.com/magefile/mage@latest && go install github.com/charmbracelet/glow@latest
RUN python3 -m pip install --upgrade pip --user && python3 -m pip install pre-commit --user
COPY files/.zshrc /home/${USERNAME}/.zshrc
COPY files/first-run-notice.txt /home/${USERNAME}/first-run-notice.txt
