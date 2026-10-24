# syntax=docker/dockerfile:1.9
#
# The Blacklight container image — the supported deployment artefact (PLAN.md §8).
#
# Build it from the repository root, because the context is the whole repo:
#
#     make docker-build
#     docker build -f deploy/Dockerfile -t blacklight:local .
#
# Three stages: the SPA, the server binary, and a runtime that carries Chromium
# for M6's PDF rendering. Everything the image needs is baked in — first boot
# makes no network request, which is the entire point of the exercise.
#
# Layer order matters and is deliberate: dependency manifests are copied before
# sources in both build stages, so editing Go code does not re-run `npm ci` and
# editing the frontend does not re-run `go mod download`.

# Debian, in every stage, because the DuckDB driver is cgo. There is no static
# musl build of it worth maintaining, so `scratch` and plain `alpine` are not
# options — see docs/deploy.md. Keeping the same distribution release in the
# build and runtime stages is what guarantees the glibc the binary was linked
# against is the glibc it finds at runtime.
#
# The base images are written out literally and pinned by digest, not built
# from ARGs: Dependabot (the docker ecosystem in .github/dependabot.yml) skips
# a FROM that uses variables, and only bumps a floating tag like trixie-slim
# when it carries a digest. Its PRs move the digest whenever the upstream image
# is rebuilt, which is how base-image security fixes reach this file. All three
# must stay on the same Debian release (trixie), for the glibc reason above.

# ─── Stage 1 · the single-page app ────────────────────────────────────────────
#
# Pinned to BUILDPLATFORM: the output is JavaScript, which is the same bytes
# whatever the target architecture, so there is no reason to run this under
# emulation when building an arm64 image on an amd64 host (or the reverse).
FROM --platform=$BUILDPLATFORM node:24.22.1-trixie-slim@sha256:43b9ba1142b8dcf33fa8f922ca1be85714adc444e0586c7cf7726daaeabc1bea AS web

WORKDIR /src/web

# Manifests only. This layer — and the install it caches — is invalidated by a
# dependency change and by nothing else.
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm,sharing=locked \
    npm ci

# `npm run build` type-checks and then runs Vite; the generated API types
# (web/src/api/schema.d.ts) are committed, so no generator runs here and the
# stage never needs api/openapi.yaml.
COPY web/ ./
RUN npm run build

# ─── Stage 2 · the server binary ──────────────────────────────────────────────
#
# Also pinned to BUILDPLATFORM, and here it is not an optimisation but the only
# thing that works: building the other architecture under QEMU segfaults the Go
# toolchain part-way through (`compile: signal: segmentation fault`), reliably
# and at a different package each time. So the Go half cross-compiles instead —
# GOARCH plus a cross C toolchain, because cgo means the C compiler has to
# target the other architecture too. Compiling natively for the host and
# cross-compiling for the other architecture take about the same time; only the
# runtime stage's `apt-get` runs emulated, and apt survives QEMU fine.
FROM --platform=$BUILDPLATFORM golang:1.27-trixie@sha256:fbb7ae40b62ff8afe31662df570d39b4f5052ac32b90dae8e60d7657a20a74f4 AS build

# Set by BuildKit, always, no default needed.
ARG BUILDARCH
ARG TARGETARCH

# g++ rather than gcc for the cross case: the DuckDB engine is a C++ static
# archive, so the link needs the target's libstdc++, which only the C++ cross
# package brings. Skipped entirely when the architectures match, which is the
# common case and where the layer stays cached.
RUN set -eu; \
    if [ "$TARGETARCH" != "$BUILDARCH" ]; then \
      case "$TARGETARCH" in \
        amd64) cross=g++-x86-64-linux-gnu ;; \
        arm64) cross=g++-aarch64-linux-gnu ;; \
        *) echo "no cross toolchain configured for TARGETARCH=$TARGETARCH" >&2; exit 1 ;; \
      esac; \
      apt-get update; \
      apt-get install -y --no-install-recommends "$cross"; \
      rm -rf /var/lib/apt/lists/*; \
    fi

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

COPY . .
# After the source copy, so the freshly built app wins over anything the context
# happened to carry. .dockerignore excludes web/dist for the same reason.
COPY --from=web /src/web/dist ./web/dist

# Version stamping. The Makefile passes real values; the defaults keep a bare
# `docker build` honest rather than claiming a version it does not have.
ARG MODULE=github.com/bryanster/blacklight
ARG VERSION=v1-dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# -tags spa is not optional: without it the binary embeds the placeholder page
# from web/dist_placeholder.go and every other check in this file still passes.
# `make build` is the reference for these flags.
#
# -trimpath and `-s -w` are the two things the image build adds: reproducible
# paths, and no DWARF or symbol table. Go tracebacks read the pclntab and are
# unaffected; a debugger attached to a container is not the plan.
# The compilation cache is keyed by architecture, so that `--platform
# linux/amd64,linux/arm64` runs the two builds in parallel instead of queueing
# on one lock. The module cache is architecture-independent and stays shared.
#
# Both binaries, in one layer: blctl is how an operator migrates, inspects and
# (from M1) creates the first user, and an image without it means reaching for
# `docker run` with a second image that has the same DuckDB version. It does not
# import web/, so it gets no `spa` tag and embeds no frontend.
ENV CGO_ENABLED=1
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,id=go-build-$TARGETARCH,sharing=locked \
    set -eu; \
    export GOARCH="$TARGETARCH"; \
    if [ "$TARGETARCH" != "$BUILDARCH" ]; then \
      case "$TARGETARCH" in \
        amd64) export CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ ;; \
        arm64) export CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ ;; \
      esac; \
    fi; \
    ldflags="-s -w \
        -X ${MODULE}/internal/version.version=${VERSION} \
        -X ${MODULE}/internal/version.commit=${COMMIT} \
        -X ${MODULE}/internal/version.buildDate=${BUILD_DATE}"; \
    go build -tags spa -trimpath -ldflags "$ldflags" -o /out/blacklight ./cmd/blacklight; \
    go build -trimpath -ldflags "$ldflags" -o /out/blctl ./cmd/blctl

# ─── Stage 3 · runtime ────────────────────────────────────────────────────────
FROM debian:trixie-slim@sha256:62bd423aeee912639ca8ee681c2b762cb3e25137e0a92b20be455898a9f034ad AS runtime

# chromium + chromium-sandbox: M6 renders PDFs by driving headless Chromium, and
#   discovering that at the end of M6 would be a packaging problem found too
#   late. chromium-sandbox is the setuid helper — installed so the sandbox can
#   stay on; docs/deploy.md covers what a host has to allow for it to work.
# fonts-*: Chromium with no fonts renders every glyph as a box. Two families,
#   ~6 MB, is the difference between a working PDF and a broken one.
# libstdc++6: the statically linked DuckDB library is C++. Chromium depends on
#   it too, but naming it here means the binary keeps running if Chromium is
#   ever dropped from the image.
# curl: the health check, and the only way to debug the API from inside.
# ca-certificates, tzdata: outbound TLS (OIDC discovery, content sync) and
#   correct timestamps.
#
# --no-install-recommends throughout; the apt lists are removed in the same
# layer so they are not carried in the image.
#
# The deletions are in the same layer as the install, because a file removed in
# a later layer is still shipped in the earlier one. What goes, and why it is
# safe: Chromium's dependency on libgl1 drags in the whole Mesa hardware GL
# stack — LLVM, Gallium, Z3, the DRI drivers — roughly 180 MB of code for
# talking to a GPU that a container does not have. Headless Chromium falls back
# to the SwiftShader software renderer bundled in /usr/lib/chromium, which is
# kept. The Vulkan validation layer is a debugging aid, and the GTK icon themes
# are for windows nobody will ever see. `chromium --headless --print-to-pdf`
# produces the same PDF with all of it gone; deploy/smoke.sh checks Chromium
# still runs, and M6's render tests are the deeper check.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates \
      chromium \
      chromium-sandbox \
      curl \
      fonts-dejavu-core \
      fonts-liberation \
      libstdc++6 \
      tzdata \
 && rm -f /usr/lib/*/libLLVM*.so* \
          /usr/lib/*/libgallium*.so* \
          /usr/lib/*/libz3.so* \
          /usr/lib/chromium/libVkLayer_khronos_validation.so \
 && rm -rf /usr/lib/*/dri /usr/share/icons /usr/share/doc \
 && rm -rf /var/lib/apt/lists/*

# A fixed uid/gid rather than whatever the base image has spare: an operator who
# bind-mounts a host directory instead of using the named volume needs a number
# to chown to, and it must not change between releases. docs/deploy.md states it.
ARG UID=10001
ARG GID=10001
#
# Not --system: that reserves uids below 1000, and a five-digit uid is the point
# — it cannot collide with a real account on a host that shares the namespace.
RUN groupadd --gid ${GID} blacklight \
 && useradd --uid ${UID} --gid ${GID} --no-create-home --no-log-init \
      --home-dir /var/lib/blacklight --shell /usr/sbin/nologin blacklight

# One directory holds everything that must survive the container: the database,
# its WAL, and uploaded evidence. Mount a volume here or lose your data — see
# docs/deploy.md. 0750, owner-only plus group, because evidence is customer data
# (internal/config uses the same mode when it creates the directory itself).
#
# Docker seeds a *named* volume from the image, ownership included, so the
# volume compose declares arrives already writable by this user. A bind mount
# does not: the host directory keeps its own ownership, which is why the number
# above is documented.
#
# No VOLUME instruction. It would make every `docker run` without -v leave an
# anonymous volume behind, and it would not save the operator who meant to
# persist their data anyway; compose declares the volume explicitly instead.
RUN install -d -o blacklight -g blacklight -m 0750 \
      /var/lib/blacklight \
      /var/lib/blacklight/evidence

# Defaults that make the image self-consistent. Everything here is overridable
# from the environment (.env.example documents each one); these are only the
# values that describe *this image's* filesystem rather than a deployment's
# policy. BLACKLIGHT_BASE_URL is one of the two required variables and has no
# sensible default — localhost is right for `docker compose up` on a laptop and
# wrong for anything else, so docs/deploy.md leads with changing it.
ENV BLACKLIGHT_ADDR=:8080 \
    BLACKLIGHT_BASE_URL=http://localhost:8080 \
    BLACKLIGHT_DB_PATH=/var/lib/blacklight/blacklight.duckdb \
    BLACKLIGHT_EVIDENCE_DIR=/var/lib/blacklight/evidence \
    BLACKLIGHT_CHROME_PATH=/usr/bin/chromium

COPY --from=build /out/blacklight /usr/local/bin/blacklight
# On PATH, so `docker compose exec blacklight blctl db info` is the whole
# command. It runs as the same user against the same volume, which is what
# makes it the supported way to touch a running deployment's database.
COPY --from=build /out/blctl /usr/local/bin/blctl
COPY --chmod=0755 deploy/entrypoint.sh /usr/local/bin/blacklight-entrypoint
COPY --chmod=0755 deploy/healthcheck.sh /usr/local/bin/blacklight-healthcheck

USER blacklight:blacklight
WORKDIR /var/lib/blacklight
EXPOSE 8080

# The health check is defined here rather than in compose.yml so that every way
# of running the image gets it — `docker run`, a compose file someone else
# wrote, an orchestrator reading the image config. compose inherits it.
#
# start-period covers migrations on a cold database; failures during it do not
# count against retries.
HEALTHCHECK --interval=15s --timeout=5s --start-period=30s --retries=3 \
    CMD ["blacklight-healthcheck"]

# The entrypoint does one job (see the script) and then execs, so the server is
# PID 1 in all but name and SIGTERM reaches it directly. CMD is the command
# rather than arguments, so `docker run <image> chromium --version` works.
ENTRYPOINT ["blacklight-entrypoint"]
CMD ["blacklight"]

# Re-declared: a build argument does not cross a stage boundary, and these must
# say the same thing as the binary's own --version. Last, so that changing them
# only invalidates the label.
ARG VERSION=v1-dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
LABEL org.opencontainers.image.title="Blacklight" \
      org.opencontainers.image.description="Self-hosted purple-team assessment platform" \
      org.opencontainers.image.source="https://github.com/bryanster/blacklight" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"
