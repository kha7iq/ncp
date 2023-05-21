# Installation

## Linux
```bash
# DEB
export NCP_VERSION="0.1.1"
wget -q https://github.com/kha7iq/ncp/releases/download/v${NCP_VERSION}/ncp_amd64.deb
sudo dpkg -i ncp_amd64.deb
# RPM
sudo rpm -i ncp_amd64.rpm
```

## Windows
```bash
scoop bucket add ncp https://github.com/kha7iq/scoop-bucket.git
scoop install ncp
```
## MacOS
```bash
brew install kha7iq/tap/ncp
```
## Manual
```bash
# Chose desired version
export NCP_VERSION="0.1.1"
wget -q https://github.com/kha7iq/ncp/releases/download/v${NCP_VERSION}/ncp_linux_amd64.tar.gz && \
tar -xf ncp_linux_amd64.tar.gz && \
chmod +x ncp && \
sudo mv ncp /usr/local/bin/.
```

## Docker

Docker container is also available on both dockerhub and github container registry.

`latest` tag will always pull the latest version available.

- Pull

```bash
docker pull khaliq/ncp:latest
```
```bash
docker pull ghcr.io/kha7iq/ncp:latest
```

- Run

```bash
docker run khaliq/ncp:latest
```

Alternatively you can head over to [release pages](https://github.com/kha7iq/ncp/releases)
and download binaries for all supported platforms.