# Usage

## Copying Files/Folders to NFS Server

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

## Writing Files on NFS Server with Specific UID and GID

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

## Copying Files/Folders from NFS Server to Local Machine

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

## File Name Truncation Option for Displaying Full Paths

**Description:**
The global flag `--truncate=false` provides a convenient way to disable file name truncation. By using this flag, the full path of the file names being transferred can be displayed without any truncation.

**Usage:**
To include the global flag `--truncate=false` in your command, follow the syntax below:

```
ncp --truncate=false from --host 192.168.0.80 --nfspath data/src 
```
**Default Value:**
By default, the `--truncate` flag is set to `true`, enabling file name truncation during file transfers. However, you have the flexibility to customize this behavior by exporting the environment variable `NCP_FILENAME_TRUNCATE` with your desired value or via flag.
