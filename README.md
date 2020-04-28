# Release_cleanup

## Rationale

You want to clean up k8s releases that have certain labels and are older than X.
It checks if the label of the deployment is not in the ignoredLabels and flags it for deletion if it's older than X days.

## Usage

```
Usage of ./cleanup:
  -age int
    	only consider releases at least this many days old (default 3)
  -excludes string
    	comma-separated list of releases to exclude
  -ignoreLabels string
    	comma-separated list of label values to ignore (default "master,preprod,dev,uat,develop")
  -label string
    	label to check against (default "branch")
  -namespace string
    	namespace to check
  -verbose
    	show branches of releases to be deleted
```
