# cloudbuild-cancel-in-progress

A Google Cloud Build helper image that automatically cancels older runs when a newer build starts.

https://hub.docker.com/r/noih/cloudbuild-cancel-in-progress

> [!WARNING]
> This only requests cancellation and does not wait for it to complete.

## Flags

| Name | Description |
| --- | --- |
| `--project-id` | Pass `$PROJECT_ID` |
| `--location` | Pass `$LOCATION` |
| `--build-id` | Pass `$BUILD_ID` |
| `--filter` | See the "About filter" section below |

### About filter

The `--filter` is transformed as follows and passed to [ListBuildsRequest](https://pkg.go.dev/cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb#ListBuildsRequest).Filter, which is an argument of [ListBuilds](https://pkg.go.dev/cloud.google.com/go/cloudbuild/apiv1/v2#Client.ListBuilds). ([cloud.google.com/go/cloudbuild](https://pkg.go.dev/cloud.google.com/go/cloudbuild/apiv1/v2) is used as the Cloud Build API client.)

```
(status="QUEUED" OR status="WORKING" OR status="PENDING") AND create_time<="{build-create-time}" AND build_id!="{build-id}" AND ({filter})
```

## How to use

### Example 1

An example of canceling older runs keyed by trigger name and branch/tag name.

```yaml
steps:
  - id: 'cancel-in-progress'
    name: 'noih/cloudbuild-cancel-in-progress:0.1.0'
    args:
      - '--project-id=$PROJECT_ID'
      - '--location=$LOCATION'
      - '--build-id=$BUILD_ID'
      - '--filter=substitutions.TRIGGER_NAME="$TRIGGER_NAME" AND substitutions.REF_NAME="$REF_NAME"'
    waitFor: ['-']
  - id: 'main'
    name: 'debian:bookworm-slim'
    args: ['bash', '-c', 'sleep 120']
    waitFor: ['-']
```

### Example 2

An example of canceling older runs keyed by trigger tag.

```yaml
steps:
  - id: 'cancel-in-progress'
    name: 'noih/cloudbuild-cancel-in-progress:0.1.0'
    args:
      - '--project-id=$PROJECT_ID'
      - '--location=$LOCATION'
      - '--build-id=$BUILD_ID'
      - '--filter=tags="cancel-group-1"'
    waitFor: ['-']
  - id: 'main'
    name: 'debian:bookworm-slim'
    args: ['bash', '-c', 'sleep 120']
    waitFor: ['-']
tags:
  - tag-1
  - cancel-group-1
```

## How to build

The build uses [ko](https://ko.build/).

```sh
KO_DOCKER_REPO="noih/cloudbuild-cancel-in-progress" ko build --bare --sbom="none" --tags="$TAG"
```
