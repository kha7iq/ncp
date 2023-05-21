# Examples

# Backup Scripts and CI/CD Pipelines

## Backup Script

In a backup script, you can utilize NCP to upload your backup files directly to an NFS server without the need for mounting the server.

Have a look at the following bash script to achive this:

```bash
#!/bin/bash

# Define the source folder and backup path
SRC_FOLDER="src"
BACKUP_FOLDER="backup"

# Create a backup with the current date appended to its name
BACKUP_FILENAME="backup_$(date +'%Y%m%d').tar.gz"
tar -czvf "${BACKUP_FOLDER}/${BACKUP_FILENAME}" "${SRC_FOLDER}"

# Upload the backup file using ncp
./ncp to --host 192.168.0.80 --nfspath data --input "${BACKUP_FOLDER}/${BACKUP_FILENAME}"
```

## CI/CD Pipelines

## .gitlab-ci.yml

This GitLab CI/CD configuration file defines a pipeline with two stages: `build` and `publish`.

The `publish` stage is triggered if the `build` stage is successful. It performs the following tasks:
- The `ncp` command is executed within the Docker container, leveraging the `docker.io/khaliq/ncp:latest` image.
- The `output.txt` file (generated in the `build` stage), uses the `ncp` command to upload the file to an NFS server.


```yaml
stages:
  - build
  - publish

build:
  stage: build
  script:
    - echo "build test artifact" > output.txt
  artifacts:
    paths:
      - output.txt
    expire_in: 2 hours

publish:
  stage: publish
  needs: ["build"]
  image: docker.io/khaliq/ncp:ncp
  script:
    - ncp to --host 192.168.0.80 --nfspath data --input output.txt

```