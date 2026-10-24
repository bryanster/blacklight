# The development container.
#
# Deliberately the same base as the image the product ships in
# (deploy/Dockerfile's build stage): same Go, same Debian release, same glibc.
# The DuckDB driver is cgo, so "it compiles in my editor" and "it compiles in
# the release image" should not be two different questions.
#
# What is *not* here is as deliberate. golangci-lint and oapi-codegen are not
# baked in: the Makefile pins their versions and installs them into ./bin, and a
# second copy in the image is a second version to drift. `make tools` is the
# postCreateCommand, and it is the same command a developer without a container
# runs.

FROM golang:1.26-trixie

# make: the Makefile is the interface to everything in this repo, and the Go
#   image does not ship it.
# sudo: passwordless for the non-root remote user — package installs and the
#   occasional root-only debug step without rebuilding the image.
# The rest: git over https and ssh, and the handful of things that make a shell
#   in here usable rather than a puzzle.
RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
	ca-certificates \
	git \
	jq \
	less \
	make \
	openssh-client \
	sudo \
	unzip \
	&& rm -rf /var/lib/apt/lists/*

# Node, taken from the official image rather than installed from a distribution
# repository, because the version has to be *exact*: .prototools, web/.nvmrc and
# web/package.json's `engines` all name 24.22.1, and so does the node stage in
# deploy/Dockerfile. A NodeSource "24.x" would drift away from that silently.
#
# Copying /usr/local wholesale would bury the Go toolchain that lives there, so
# only Node's own files come across, and the npm/npx entry points are relinked.
COPY --from=node:24.22.1-trixie-slim /usr/local/lib/node_modules /usr/local/lib/node_modules
COPY --from=node:24.22.1-trixie-slim /usr/local/bin/node /usr/local/bin/node
RUN ln -sf /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
	&& ln -sf /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx \
	&& ln -sf /usr/local/lib/node_modules/corepack/dist/corepack.js /usr/local/bin/corepack \
	&& node --version && npm --version

RUN npm install -g @anthropic-ai/claude-code

# Prime Agent. The installer resolves the latest release, verifies its checksum
# and hands the tarball to `npm install -g`, so the command lands in
# /usr/local/bin next to the others. Two of its steps are worth pinning down
# rather than leaving to the defaults:
#   PLAIN — the installer draws an animated full-screen TUI when stdout is a
#     terminal. It is not, during a build, but the variable says so explicitly
#     instead of relying on that.
#   BOOTSTRAP_KERNEL=0 — it otherwise installs uv, Python 3.11 and an ipykernel
#     at this point. That runtime lives in $HOME, which is still root's here;
#     baked now it would be a root-owned tree the vscode user cannot write.
#     Left off, prime-agent prepares it on first use, in the right home.
# Node is already on the PATH and newer than the installer's floor, so its own
# Node bootstrap (and the sudo that goes with it) never runs.
RUN curl -fsSL https://app.primeintellect.ai/prime-agent/install.sh \
	| env PRIME_AGENT_INSTALLER_PLAIN=1 PRIME_AGENT_BOOTSTRAP_KERNEL_ON_INSTALL=0 sh \
	&& prime-agent --version

RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Non-root development user. remoteUser in devcontainer.json matches this name;
# the docker-outside-of-docker feature then adds it to the docker group so
# `make docker-build` works without sudo. updateRemoteUserUID (the default)
# rewrites the uid/gid at create time to match the host user, so bind-mounted
# sources stay writable.
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000
RUN groupadd --gid ${USER_GID} ${USERNAME} \
	&& useradd --uid ${USER_UID} --gid ${USER_GID} -m -s /bin/bash ${USERNAME} \
	&& echo "${USERNAME} ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
	&& chmod 0440 /etc/sudoers.d/${USERNAME} \
	&& mkdir -p /go/pkg/mod /go/bin \
	&& chown -R ${USERNAME}:${USERNAME} /go /home/${USERNAME}

# Oh My Pi (omp). PI_INSTALL_DIR overrides the installer default of
# $HOME/.local/bin so the binary lands on the shared PATH for the non-root
# user rather than under /root. --binary skips the bun source path; we only
# need the released agent binary in this image.
ENV PI_INSTALL_DIR=/usr/local/bin
RUN curl -fsSL https://omp.sh/install | sh -s -- --binary \
	&& omp --version

# Not set in the image: BLACKLIGHT_* configuration. The server reads it from the
# environment and refuses anything it cannot use, so a container-wide default
# would be a value nobody chose, quietly in force. Copy .env.example to .env.
USER ${USERNAME}
