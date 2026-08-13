FROM mcr.microsoft.com/devcontainers/base:dev-ubuntu24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    bash-completion \
    build-essential \
    ca-certificates \
    curl \
    git \
    jq \
    less \
    nodejs \
    npm \
    python3 \
    python3-pip \
    python3-venv \
    python3-full \
    ripgrep \
    tmux \
    unzip \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

RUN ln -sf /usr/bin/python3 /usr/local/bin/python

RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

RUN npm install -g @openai/codex @anthropic-ai/claude-code

# Go toolchain — pukk-test builds a single cross-compiled executable
# (Linux devcontainer -> Windows binary via GOOS=windows go build).
ARG GO_VERSION=1.26.5
RUN ARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

COPY ./.devcontainer/init-shell.sh /etc/profile.d/init-shell.sh
RUN chmod +x /etc/profile.d/init-shell.sh