package main

import (
	"encoding/json"
	"strings"
	"time"
)

const HelmTimeLayout = "Tue Oct 22 22:45:51 2019"

type RawJson map[string]interface{}
type DeployDates map[string]time.Time

func GetLabels(b []byte) map[string]string {
	result := RawJson{}
	labels := map[string]string{}
	err := json.Unmarshal(b, &result)
	if err != nil {
		return map[string]string{"": ""}
	}

	items := result["items"].([]interface{})
	for _, v := range items {
		item := v.(map[string]interface{})
		metadata := item["metadata"].(map[string]interface{})
		for k, v := range metadata["labels"].(map[string]interface{}) {
			labels[k] = v.(string)
		}
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
		stime := strings.TrimSpace(splittedString[2])
		time, _ := time.Parse(HelmTimeLayout, stime)
		result[name] = time
	}
	return result
}
