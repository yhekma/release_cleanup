package main

import (
	"io/ioutil"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	t.Run("see if we can parse output correctly", func(t *testing.T) {
		fBytes, _ := ioutil.ReadFile("k_output.json")
		result := GetLabels(fBytes)
		if result["app"] == "" {
			t.Fatal("Could not parse json")
		}
	})

	t.Run("test if we can get dates from helm output", func(t *testing.T) {
		helmData := []byte(`
		NAME              	REVISION	UPDATED                 	STATUS  	CHART                   	NAMESPACE
		track-if2nova-grpc	15      	Tue Oct 22 22:45:51 2019	DEPLOYED	track-if2nova-0.2.4     	mytnt2
		uk-booking-service	21      	Thu Oct 17 09:13:16 2019	DEPLOYED	uk-booking-service-0.1.0	mytnt2
`)
		trackTime, _ := time.Parse(HelmTimeLayout, "Tue Oct 22 22:45:51 2019")
		ukBookingTime, _ := time.Parse(HelmTimeLayout, "Thu Oct 17 09:13:16 2019")
		dates := map[string]time.Time{
			"track-if2nova-grpc": trackTime,
			"uk-booking-service": ukBookingTime,
		}
		result := GetDeployDates(helmData)
		for k, v := range dates {
			if result[k] != v {
				t.Errorf("got incorrect time for %s, want %s, got %s", k, result[k], v)
			}
		}
	})
}
