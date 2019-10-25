package main

import (
	"encoding/json"
	"strings"
	"time"
)

const HelmTimeLayout = "Tue Oct 22 22:45:51 2019"

type RawJson map[string]interface{}
type DeployDates map[string]time.Time
type ReleaseLabels map[string]map[string]string

func GetLabels(b []byte, filter string) ReleaseLabels {
	// filter is key that needs to exist in labels
	// Return will look like
	// {m3db: {"app": "m3db-node", "controller-revision": "bla" }}
	rjson := RawJson{}
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
