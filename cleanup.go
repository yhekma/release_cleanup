package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

const HelmTimeLayout = "Mon Jan 02 15:04:05 2006"

type rawJson map[string]interface{}
type DeployDates map[string]time.Time

func GetMatchingPods(b []byte, filter string) []string {
	rjson := rawJson{}
	var result []string
	err := json.Unmarshal(b, &rjson)
	if err != nil {
		log.Fatal(err)
	}

	items := rjson["items"].([]interface{})
	for _, v := range items {
		item := v.(map[string]interface{})
		metadata := item["metadata"].(map[string]interface{})
		labels := metadata["labels"].(map[string]interface{})
		if labels[filter] == nil {
			continue
		}
		result = append(result, labels["release"].(string))
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
	cmd := exec.Command("kubectl", "get", "pods", "-o", "json", "-n", namespace)
	result := getOutput(cmd)
	return result
}

func GetHelmOutput() []byte {
	cmd := exec.Command("helm", "list", "--all")
	result := getOutput(cmd)
	return result
}

func Contains(s []string, item string) bool {
	for _, i := range s {
		if i == item {
			return true
		}
	}
	return false
}

func main() {
	filter := flag.String("filter", "tbc", "only look for pods with this label set")
	age := flag.Int("age", 3, "only consider releases at least this many days old")
	flag.Parse()
	kubeOutput := GetKubeOutput("mytnt2")
	helmOutput := GetHelmOutput()
	deployDates := GetDeployDates(helmOutput)
	matchingPods := GetMatchingPods(kubeOutput, *filter)
	releasesToBeConsidered := GetOlderReleases(deployDates, *age)
	for _, release := range releasesToBeConsidered {
		if Contains(matchingPods, release) {
			fmt.Println(release)
		}
	}
}
