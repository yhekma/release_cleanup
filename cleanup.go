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
	"regexp"
	"strings"
	"time"
)

const HelmTimeLayout = "Mon Jan 2 15:04:05 2006"

type kubeResponse struct {
	Items []struct {
		Metadata struct {
			Labels map[string]interface{} `json:"labels"`
		} `json:"metadata"`
	} `json:"items"`
}

type DeployDates map[string]time.Time

func GetMatchingReleases(b []byte, ignoreBranches []string, excludes []string) []string {
	response := kubeResponse{}
	var result []string
	err := json.Unmarshal(b, &response)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range response.Items {
		labels := v.Metadata.Labels
		release := labels["release"].(string)

		if labels["branch"] == nil {
			continue
		}

		if Contains(ignoreBranches, labels["branch"].(string)) == true {
			continue
		}

		if Contains(excludes, release) {
			continue
		}

		matched, _ := regexp.MatchString(`(uat|preprod|dev|develop|test)$`, release)
		if matched {
			continue
		}

		result = append(result, release)
	}
	return result
}

func GetDeployDates(b []byte) DeployDates {
	result := DeployDates{}
	splittedStringLines := strings.Split(string(b), "\n")
	for _, l := range splittedStringLines {
		if l == "" {
			continue
		}
		splittedString := strings.Split(l, "\t")
		if len(splittedString) < 6 {
			continue
		}
		name := strings.TrimSpace(splittedString[0])
		if name == "NAME" {
			continue
		}
		stringTime := strings.TrimSpace(splittedString[2])
		parsedTime, _ := time.Parse(HelmTimeLayout, stringTime)
		result[name] = parsedTime
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
	cmd := exec.Command("kubectl", "get", "deployments", "-o", "json", "-n", namespace)
	result := getOutput(cmd)
	return result
}

func GetHelmOutput() []byte {
	cmd := exec.Command("helm", "list", "--all")
	result := getOutput(cmd)
	return result
}

func deleteReleases(releases []string, pretend bool) []byte {
	var cmd *exec.Cmd
	if pretend {
		cmd = exec.Command("echo", "pretend", "helm", "delete", "--purge", strings.Join(releases, " "))
	} else {
		cmd = exec.Command("echo", "helm", "delete", "--purge", strings.Join(releases, " "))
	}
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
	namespace := flag.String("namespace", "mytnt2", "namespace to check")
	pretend := flag.Bool("pretend", false, "run in pretend mode")
	exclude := flag.String("excludes", "", "comma-separated list of releases to exclude")

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
	ignoreBranches := strings.Split(*fignoreBranches, ",")
	excludes := strings.Split(*exclude, ",")
	kubeOutput := GetKubeOutput(*namespace)
	helmOutput := GetHelmOutput()
	deployDates := GetDeployDates(helmOutput)
	matchingReleases := GetMatchingReleases(kubeOutput, ignoreBranches, excludes)
	oldReleases := GetOlderReleases(deployDates, *age)
	releasesToBeDeleted := intersect(oldReleases, matchingReleases)
	result := deleteReleases(releasesToBeDeleted, *pretend)
	fmt.Println(string(result))
}
