package main

import (
	"io/ioutil"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	trackTime, _ := time.Parse(HelmTimeLayout, "Tue Oct 01 22:45:51 2019")
	ukBookingTime, _ := time.Parse(HelmTimeLayout, "Thu Oct 05 09:13:16 2019")
	dates := DeployDates{
		"track-if2nova-grpc": trackTime,
		"uk-booking-service": ukBookingTime,
	}

	t.Run("see if we can find matching pods", func(t *testing.T) {
		fBytes, _ := ioutil.ReadFile("k_output.json")
		result := GetMatchingReleases(fBytes, "app")
		if !Contains(result, "m3db") {
			t.Errorf("expected item ('m3db') does not appear in output")
		}
	})

	t.Run("see if we can find non matching pods", func(t *testing.T) {
		fBytes, _ := ioutil.ReadFile("k_output.json")
		result := GetMatchingReleases(fBytes, "xxx")
		if len(result) > 0 {
			t.Errorf("found too many matching pods")
		}
	})

	t.Run("test if we can get dates from helm output", func(t *testing.T) {
		helmData := []byte(`
NAME              	REVISION	UPDATED                 	STATUS  	CHART                   	NAMESPACE
track-if2nova-grpc	15      	Tue Oct 01 22:45:51 2019	DEPLOYED	track-if2nova-0.2.4     	mytnt2
uk-booking-service	21      	Thu Oct 05 09:13:16 2019	DEPLOYED	uk-booking-service-0.1.0	mytnt2
`)
		result := GetDeployDates(helmData)
		for k, v := range dates {
			if result[k] != v {
				t.Errorf("got incorrect time for %s, want %s, got %s", k, result[k], v)
			}
		}
	})

	t.Run("get matching releases", func(t *testing.T) {
		resultTrue := GetOlderReleases(dates, 3)
		resultFalse := GetOlderReleases(dates, 99999)
		if len(resultTrue) != 2 {
			t.Errorf("could not find older releases")
		}
		if len(resultFalse) != 0 {
			t.Errorf("could not find newer releases")
		}
	})
}
