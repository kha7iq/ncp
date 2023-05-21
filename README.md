<h2 align="center">
  <p align="center"><img width=30% src="./.github/img/logo.png"></p>
</h2>
<p align="center">
  <img alt="GitHub Build Status" src="https://img.shields.io/github/actions/workflow/status/kha7iq/ncp/build.yml?label=Build">
   <a href="https://github.com/kha7iq/ncp/releases">
   <img alt="Release" src="https://img.shields.io/github/v/release/kha7iq/ncp?label=Release">
   <a href="https://goreportcard.com/report/github.com/kha7iq/ncp">
   <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/kha7iq/ncp">
   <a href="#">
   <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/kha7iq/ncp">
   <a href="https://github.com/kha7iq/ncp/issues">
   <img alt="GitHub issues" src="https://img.shields.io/github/issues/kha7iq/ncp?style=flat-square&logo=github&logoColor=white">
   <a href="https://github.com/kha7iq/ncp/blob/master/LICENSE.md">
   <img alt="License" src="https://img.shields.io/github/license/kha7iq/ncp">
</p>

<p align="center">
  <a href="#installation">Installation</a> •
  <a href="#features">Features</a> •
  <a href="#usage">Usage</a> •
  <a href="#contributing">Contributing</a> •
</p>

# NCP File Transfer Utility (NFSv3)

NCP is a file transfer utility that enables efficient copying of files to and from an NFS server. It offers a convenient way to transfer files between your local machine and an NFS server, supporting both upload and download operations.

## Features

- Easy file transfer to and from an NFS server
- Support for upload and download operations
- Multi-architecture binaries available for installation (e.g., .deb, apk, rpm)
- Compatible with Windows and macOS operating systems
- Option to specify UID and GID for write operations using a global flag

<img style="border:0.5px solid silver;" alt="NCP" src="./.github/img/ncp.gif" width="800" />



## Installation


<details>
    <summary>DEB & RPM</summary>

```bash
# DEB
sudo dpkg -i ncp_amd64.deb
# RPM
sudo rpm -i ncp_amd64.rpm
```
</details>

<details>
    <summary>Windows</summary>

```bash
scoop bucket add ncp https://github.com/kha7iq/scoop-bucket.git
scoop install ncp
```
</details>

<details>
    <summary>MacOS</summary>

```bash
brew install kha7iq/tap/ncp
```
</details>

<details>
    <summary>Manual</summary>

```bash
# Chose desired version
export NCP_VERSION="0.1.1"
wget -q https://github.com/kha7iq/ncp/releases/download/v${NCP_VERSION}/ncp_linux_amd64.tar.gz && \
tar -xf ncp_linux_amd64.tar.gz && \
chmod +x ncp && \
sudo mv ncp /usr/local/bin/.
```
</details>

Alternatively you can head over to [release pages](https://github.com/kha7iq/ncp/releases)
and download binaries for all supported platforms.

## Docker

Docker container is also available on both dockerhub and github container registry.

`latest` tag will always pull the latest version available.
<details>
    <summary>Docker</summary>

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
</details>

## Usage

### Copying Files/Folders to NFS Server

To copy a file or folder to an NFS server, use the following command:
```bash
ncp to --host <NFS_SERVER> --nfspath <NFS_PATH> --input <LOCAL_PATH>
```
Replace the placeholders `<NFS_SERVER>`, `<NFS_PATH>`, and `<LOCAL_PATH>` with the appropriate values:

- `<NFS_SERVER>`: The IP address or hostname of the NFS server.
- `<NFS_PATH>`: The path on the NFS server where the files/folder will be copied.
- `<LOCAL_PATH>`: The local path to the file or folder you want to copy.

For example, to copy the `_local/src` folder to the NFS server with the IP address `192.168.0.80` and the NFS path `data`, use the following command:
```bash
ncp to --host 192.168.0.80 --nfspath data --input _local/src
```

### Writing Files on NFS Server with Specific UID and GID

If you want to write files on the NFS server with a specific UID and GID, you can provide the global flags `--uid` and `--gid` in the command. Use the following format:
```bash
ncp --uid <UID> --gid <GID> to --host <NFS_SERVER> --nfspath <NFS_PATH> --input <LOCAL_PATH>
```
Replace `<UID>` and `<GID>` with the desired UID and GID values. The other placeholders have the same meaning as explained in the previous section.

For example, to copy the `_local/src` folder to the NFS server with the IP address `192.168.0.80` and the NFS path `data`, while setting the UID to `1000` and GID to `1000`, use the following command:
```bash
ncp --uid 1000 --gid 1000 to --host 192.168.0.80 --nfspath data --input _local/src
```
If no UID or GID is provided, ncp will use a default value of 0.

### Copying Files/Folders from NFS Server to Local Machine

To copy a file or folder from the NFS server to your local machine, use the following command:
```bash
ncp from --host <NFS_SERVER> --nfspath <NFS_PATH>
```

Replace `<NFS_SERVER>` and `<NFS_PATH>` with the appropriate values:

- `<NFS_SERVER>`: The IP address or hostname of the NFS server.
- `<NFS_PATH>`: The path on the NFS server from where the files/folder will be copied.

For example, to copy the `src` folder recursively from the NFS server with the IP address `192.168.0.80` and the NFS path `data` to the current folder on your local machine, use the following command:
```bash
ncp from --host 192.168.0.80 --nfspath data/src
```

## Contributing

Contributions, issues and feature requests are welcome!<br/>Feel free to check
[issues page](https://github.com/kha7iq/ncp/issues). You can also take a look
at the [contributing guide](https://github.com/kha7iq/ncp/blob/master/CONTRIBUTING.md).

## Issues

If you encounter any problems or have suggestions for improvements, please [open an issue](https://github.com/username/repo/issues) on GitHub.

---
## License

NCP is licensed under the MIT License. Please note that it may use third-party libraries that have their own separate licenses. Refer to the individual licenses of those libraries for more information.