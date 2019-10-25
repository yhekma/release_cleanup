package main

import (
	"encoding/json"
	"strings"
	"time"
)

const HelmTimeLayout = "Mon Jan 02 15:04:05 2006"

type rawJson map[string]interface{}
type DeployDates map[string]time.Time
type ReleaseLabels map[string]map[string]string

func GetLabels(b []byte, filter string) ReleaseLabels {
	// filter is key that needs to exist in labels
	// Return will look like
	// {m3db: {"app": "m3db-node", "controller-revision": "bla" }}
	rjson := rawJson{}
	labels := ReleaseLabels{}
	_ = json.Unmarshal(b, &rjson)

	items := rjson["items"].([]interface{})
	for _, v := range items {
		item := v.(map[string]interface{})
		metadata := item["metadata"].(map[string]interface{})
		rlabels := metadata["labels"].(map[string]interface{})
		if rlabels[filter] == nil {
			continue
		}
		release := rlabels["release"].(string)
		foundLabels := map[string]string{}
		for k, v := range rlabels {
			foundLabels[k] = v.(string)
		}
		labels[release] = foundLabels
	}

	return labels
}

func GetDeployDates(b []byte) DeployDates {
	result := DeployDates{}
	splittedStringLines := strings.Split(string(b), "\n")
	for _, l := range splittedStringLines {
		if l == "" {
			continue
		}
		splittedString := strings.Split(l, "\t")
		name := strings.TrimSpace(splittedString[2])
		if name == "NAME" {
			continue
		}
		stringTime := strings.TrimSpace(splittedString[4])
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
