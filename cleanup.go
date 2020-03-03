package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

const HelmTimeLayout = "Mon Jan 2 15:04:05 2006"

type kubeResponse struct {
	Items []struct {
		Metadata struct {
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
	} `json:"items"`
}

type helmResponse struct {
	Releases []struct {
		Name    string `json:"Name"`
		Updated string `json:"Updated"`
	} `json:"Releases"`
}

type DeployDates map[string]time.Time

func GetMatchingReleases(b []byte, ignoreBranches []string, excludes []string) map[string]string {
	response := kubeResponse{}
	result := make(map[string]string)
	err := json.Unmarshal(b, &response)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range response.Items {
		labels := v.Metadata.Labels

		// Disregard deployments that don't have "branch" or "release" in the labels
		if _, ok := labels["release"]; !ok {
			continue
		}
		if _, ok := labels["branch"]; !ok {
			continue
		}

		switch {
		case Contains(ignoreBranches, labels["branch"]) == true:
			continue
		case Contains(excludes, labels["release"]):
			continue
		default:
			result[labels["release"]] = labels["branch"]
		}
	}
	return result
}

func GetDeployDates(b []byte) DeployDates {
	response := helmResponse{}
	err := json.Unmarshal(b, &response)
	if err != nil {
		log.Fatal(err)
	}
	result := DeployDates{}
	for _, release := range response.Releases {
		parsedTime, _ := time.Parse(HelmTimeLayout, release.Updated)
		result[release.Name] = parsedTime
	}
	return result
}

func GetOlderReleases(dates DeployDates, days int) []string {
	now := time.Now()
	var result []string
	for deploy, date := range dates {
		targetDate := date.AddDate(0, 0, days)
		if now.After(targetDate) {
			result = append(result, deploy)
		}
	}

	return result
}

func getOutput(cmd *exec.Cmd) []byte {
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return output.Bytes()
}

func GetKubeOutput(namespace string) []byte {
	var cmd *exec.Cmd
	if namespace == "" {
		cmd = exec.Command("kubectl", "get", "deployments", "-o", "json", "--all-namespaces")
	} else {
		cmd = exec.Command("kubectl", "get", "deployments", "-o", "json", "-n", namespace)
	}
	result := getOutput(cmd)
	return result
}

func GetHelmOutput() []byte {
	// Getting helm output on helm 2.13.1 with --all gives less output than without --all for some reason.....
	cmd := exec.Command("helm", "list", "--output", "json")
	result := getOutput(cmd)
	return result
}

func deleteReleases(releases []string) []byte {
	cmd := exec.Command("echo", "helm", "delete", "--purge", strings.Join(releases, " "))
	return getOutput(cmd)
}

func Contains(s []string, item string) bool {
	for _, i := range s {
		if i == item {
			return true
		}
	}
	return false
}

// This is fairly slow and naive, but we are not dealing with large slices here (think 100s tops), so readability before scalability
func intersect(slice1, slice2 []string) []string {
	var result []string
	for _, i := range slice1 {
		if Contains(slice2, i) {
			result = append(result, i)
		}
	}
	return result
}

func main() {
	fignoreBranches := flag.String("ignoreBranches", "master,preprod,dev,uat,develop", "comma-separated list of branches to ignore")
	age := flag.Int("age", 3, "only consider releases at least this many days old")
	namespace := flag.String("namespace", "", "namespace to check")
	exclude := flag.String("excludes", "", "comma-separated list of releases to exclude")
	verbose := flag.Bool("verbose", false, "show branches of releases to be deleted")

	// Test if we can read provided kubeconfig
	kubeConfig := os.Getenv("KUBECONFIG")
	_, err := os.Open(kubeConfig)
	if err != nil {
		// See if we can read the default kubeconfig
		homeDir := os.Getenv("HOME")
		_, err := os.Open(path.Join(homeDir, ".kube", "config"))
		// Error out
		if err != nil {
			log.Fatalf("Could not read kubeconfig:\n\t%v\n", err)
		}
	}

	flag.Parse()

	var (
		kubeOutput []byte
		helmOutput []byte
	)
	var (
		helmwg sync.WaitGroup
		kubewg sync.WaitGroup
	)
	helmwg.Add(1)
	kubewg.Add(1)

	ignoreBranches := strings.Split(*fignoreBranches, ",")
	excludes := strings.Split(*exclude, ",")

	go func(ns string, w *sync.WaitGroup) {
		kubeOutput = GetKubeOutput(ns)
		defer w.Done()
	}(*namespace, &kubewg)
	go func(w *sync.WaitGroup) {
		helmOutput = GetHelmOutput()
		defer w.Done()
	}(&helmwg)
	helmwg.Wait()
	kubewg.Wait()

	deployDates := GetDeployDates(helmOutput)
	matchingReleases := GetMatchingReleases(kubeOutput, ignoreBranches, excludes)
	oldReleases := GetOlderReleases(deployDates, *age)

	// Get keys from matchingReleases map
	matchingReleasesSlice := make([]string, 0, len(matchingReleases))
	for k := range matchingReleases {
		matchingReleasesSlice = append(matchingReleasesSlice, k)
	}

	releasesToBeDeleted := intersect(oldReleases, matchingReleasesSlice)
	result := deleteReleases(releasesToBeDeleted)
	fmt.Println(string(result))
	if *verbose {
		// Get longest string
		var releaseLength int
		var branchLength int
		for _, releaseString := range releasesToBeDeleted {
			if len(releaseString) > releaseLength {
				releaseLength = len(releaseString)
			}
		}
		for _, release := range releasesToBeDeleted {
			if len(matchingReleases[release]) > branchLength {
				branchLength = len(matchingReleases[release])
			}
		}

		fmt.Printf("\n%-*s BRANCH %*s\n", releaseLength+5, "RELEASE", branchLength+28, "DEPLOY DATE (dd-mm-yyyyy hh:mm)")
		for _, release := range releasesToBeDeleted {
			fmt.Printf("%-*s  --  %-*s -- %s\n", releaseLength, release, branchLength, matchingReleases[release], deployDates[release].Format("02-01-2006 15:04"))
		}
	}
}
