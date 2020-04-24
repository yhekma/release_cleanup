# Release_cleanup

## Rationale

You want to clean up k8s releases that have certain labels and are older than X

## Usage

```
Usage of ./cleanup:
  -age int
    	only consider releases at least this many days old (default 3)
  -excludes string
    	comma-separated list of releases to exclude
  -ignoreBranches string
    	comma-separated list of branches to ignore (default "master,preprod,dev,uat,develop")
  -namespace string
    	namespace to check
  -verbose
    	show branches of releases to be deleted```
