package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/google-api-go-client/googleapi"
	"github.com/google/google-api-go-client/storage/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var c *http.Client
var s *storage.Service

func init() {
	var err error
	c, err = google.DefaultClient(oauth2.NoContext, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		panic("No client")
	}

	s, err = storage.New(c)
	if err != nil {
		panic("No service")
	}
}

const (
	bucket   = "gce-ilb-load"
	outputjs = "data.json"
)

func main() {
	fmt.Printf("Processing jobs from %v\n", bucket)
	jobs, err := listJobs(bucket, "")
	if err != nil {
		log.Fatalf("error %s", err.Error)
	}

	lts := make([]*LoadTest, len(jobs))
	resc, errc := make(chan *LoadTest), make(chan error)

	for _, j := range jobs {
		go func(job string) {
			lt, err := getLoadTestForJob(job)
			if err != nil {
				errc <- err
			}
			resc <- lt
		}(j)
	}

	for i := 0; i < len(jobs); i++ {
		select {
		case res := <-resc:
			lts[i] = res
		case err := <-errc:
			log.Fatalf("Error: %s", err.Error())
		}
	}

	outbytes, err := json.Marshal(lts)
	if err != nil {
		log.Fatalf("error marshaling json: %s", err.Error)
	}

	outs := "var testdata=" + string(outbytes)
	err = ioutil.WriteFile(outputjs, []byte(outs), 0644)
	if err != nil {
		log.Fatalf("error writing file: %s", err.Error)
	}
}

func listJobs(bucket string, prefix string) ([]string, error) {
	objects, err := listObjects(bucket, "/", "")
	if err != nil {
		return nil, err
	}
	return objects.Prefixes, nil
}

func getLoadTestForJob(jobPrefix string) (*LoadTest, error) {
	loadtest := &LoadTest{}
	os, err := listObjects(bucket, "", jobPrefix)
	if err != nil {
		return loadtest, err
	}
	for _, o := range os.Items {
		if strings.HasSuffix(o.Name, "job.json") {
			oc, err := getObjectContents(o)
			if err != nil {
				return loadtest, err
			}
			err = json.Unmarshal(oc, loadtest)
			if err != nil {
				return loadtest, err
			}
			break
		}
	}

	// Append all test results
	allResults, err := getTestResultsForJob(jobPrefix)
	if err != nil {
		return loadtest, err
	}
	loadtest.AddResults(allResults)
	return loadtest, nil
}

func getTestResultsForJob(jobPrefix string) ([]TestResult, error) {
	os, err := listObjects(bucket, "", jobPrefix)
	if err != nil {
		return nil, err
	}

	results := make([]TestResult, len(os.Items)-1)
	resc, errc := make(chan TestResult), make(chan error)

	for _, o := range os.Items {
		go func(obj *storage.Object) {
			if !strings.HasSuffix(obj.Name, "job.json") {
				tr, err := getTestResult(obj)
				if err != nil {
					errc <- err
				}
				resc <- tr
			}
		}(o)
	}

	for i := 0; i < len(os.Items)-1; i++ {
		select {
		case res := <-resc:
			results[i] = res
		case err := <-errc:
			return results, err
		}
	}

	return results, nil
}

func getTestResult(o *storage.Object) (TestResult, error) {
	tr := TestResult{}
	oc, err := getObjectContents(o)
	if err != nil {
		return tr, err
	}
	err = json.Unmarshal(oc, &tr)
	if err != nil {
		return tr, err
	}
	return tr, nil
}

func getObjectContents(o *storage.Object) ([]byte, error) {
	urls := "https://www.googleapis.com/download/storage/v1/b/{BUCKET}/o/{OBJECT}?alt=media"

	req, _ := http.NewRequest("GET", urls, nil)
	req.URL.Path = strings.Replace(req.URL.Path, "{BUCKET}", url.QueryEscape(o.Bucket), 1)
	req.URL.Path = strings.Replace(req.URL.Path, "{OBJECT}", url.QueryEscape(o.Name), 1)
	googleapi.SetOpaque(req.URL)

	r, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func listObjects(bucket string, delimeter string, prefix string) (*storage.Objects, error) {
	os := storage.NewObjectsService(s)
	objects, err := os.List(bucket).MaxResults(10000).Prefix(prefix).Delimiter(delimeter).Do()
	if err != nil {
		return objects, err
	}
	for objects.NextPageToken != "" {
		page, err := os.List(bucket).MaxResults(10000).Prefix(prefix).Delimiter(delimeter).PageToken(objects.NextPageToken).Do()
		if err != nil {
			return page, err
		}
		objects.Items = append(objects.Items, page.Items...)
		objects.NextPageToken = page.NextPageToken
	}
	return objects, nil
}
