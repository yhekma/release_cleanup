# Release_cleanup

## Rationale

You want to clean up k8s releases that have certain labels and are older than X.
It checks if the label of the deployment is not in the ignoredLabels and flags it for deletion if it's older than X days.

## How it works

The program goes through all the deployments in a given namespace (or all by default) and checks for a given label and value of that label and which helm releases this deployment belongs to via the `release` label.
If the value matches and the age of the helm releases is more than the given nr of days, it's flagged for deletion.

## Usage

```
Usage of ./release_cleanup:
  -age int
    	only consider releases at least this many days old (default 3)
  -excludes string
    	comma-separated list of releases to exclude
  -ignoreLabels string
    	comma-separated list of label values to ignore (default "master,preprod,dev,uat,develop")
  -label string
    	label to check against, deployments without this label will be ignored (default "branch")
  -namespace string
    	namespace to check, defaults to all namespaces
  -verbose
    	show branches of releases to be deleted
```

Note that when running inside docker, the home dir is `/app` so any kubeconfig should be mounted in there (or specified with `-e KUBECONFIG=<path>` to docker)

## Examples

Clean releases older than 2 days but not when label value is `master`, `preprod`, `dev`, `uat`, `develop`:

```docker run --rm -v $HOME/.kube:/app/.kube:ro local/release_cleanup -age 2 -verbose```

Clean reseases older than 5 days except releases `foo` and `bar`

`docker run --rm -v $HOME/.kube:/app/.kube:ro local/release_cleanup -age 5 -excludes foo,bar -label release`