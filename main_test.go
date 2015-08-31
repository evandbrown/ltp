package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestListObjects(t *testing.T) {
	objects, err := listObjects("gce-ilb-load", "", "")
	if err != nil {
		t.Errorf("error %s", err)
	}

	for _, o := range objects.Items {
		fmt.Println(o.MediaLink)
	}

}

func TestPrintObjects(t *testing.T) {
	objects, err := listObjects("gce-ilb-load", "", "")
	if err != nil {
		t.Errorf("error %s", err)
	}

	for _, o := range objects.Items {
		c, err := getObjectContents(o)
		if err != nil {
			t.Errorf("error %s", err)
		}
		fmt.Println(string(c))
		break
	}
}

func TestListJobs(t *testing.T) {
	jobs, err := listJobs("gce-ilb-load", "")
	if err != nil {
		t.Errorf("error %s", err)
	}
	for _, j := range jobs {
		fmt.Println(j)
	}
}

func TestListObjectsByJob(t *testing.T) {
	jobs, err := listJobs("gce-ilb-load", "")
	if err != nil {
		t.Errorf("error %s", err)
	}
	for _, j := range jobs {
		fmt.Println(j)
		os, err := listObjects("gce-ilb-load", "", j)
		if err != nil {
			t.Errorf("error %s", err)
		}
		for _, o := range os.Items {
			fmt.Println("-- " + o.Name)
		}
	}
}

func TestMarshalJobResults(t *testing.T) {
	jobs, err := listJobs("gce-ilb-load", "")
	if err != nil {
		t.Errorf("error %s", err)
	}
	for _, j := range jobs {
		os, err := listObjects("gce-ilb-load", "", j)
		if err != nil {
			t.Errorf("error %s", err)
		}
		for _, o := range os.Items {
			if !strings.HasSuffix(o.Name, "job.json") {
				tr, err := getTestResult(o)
				if err != nil {
					t.Errorf("error %s", err)
				}
				fmt.Println(tr)
				return
			}
		}
	}
}

func TestGetAllResultsForJob(t *testing.T) {
	jobs, err := listJobs("gce-ilb-load", "")
	if err != nil {
		t.Errorf("error %s", err)
	}
	j := jobs[0]
	results, err := getTestResultsForJob(j)
	if err != nil {
		t.Errorf("err %s", err)
	}
	fmt.Printf("got %v jobs\n", len(results))
}

func TestResultAggregation(t *testing.T) {
	lt := &LoadTest{}
	lt.AddResults([]TestResult{{LatencyMax: 1, RequestsMax: 2}, {LatencyMax: 1, RequestsMax: 2}})
	if lt.ResultsAveraged.LatencyMax != 1 {
		t.Errorf("invalid LatencyMax (%v)", lt.ResultsAveraged.LatencyMax)
	}
	if lt.ResultsAveraged.RequestsMax != 4 {
		t.Errorf("invalid RequestsMax (%v)", lt.ResultsAveraged.RequestsMax)
	}
}

func TestLoadAllJobs(t *testing.T) {
	jobs, err := listJobs(bucket, "")
	if err != nil {
		t.Errorf("error %s", err)
	}
	for _, j := range jobs {
		lt, err := getLoadTestForJob(j)
		if err != nil {
			t.Errorf("error %s", err)
		}
		fmt.Printf("Mean Latency for job: %v\n", lt.ResultsAveraged.LatencyMean)
		fmt.Printf("Requests for job: %v\n", lt.ResultsAveraged.SummaryRequests)
	}
}
