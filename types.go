package main

import (
	"reflect"
)

type LoadTest struct {
	Timestamp         string     `json:"timestamp"`
	ConnsPerWorker    string     `json:"conns_per_worker"`
	Duration          string     `json:"duration"`
	DeploymentID      string     `json:"deployment_id"`
	Workers           string     `json:"workers"`
	Target            string     `json:"target"`
	ThreadsPerWorker  string     `json:"threads_per_worker"`
	AggregatedResults TestResult `json:"results_averaged"`
}

func (lt *LoadTest) AddResults(results []TestResult) {
	for _, r := range results {
		resultType := reflect.ValueOf(&r).Elem()
		typeOfT := resultType.Type()
		for i := 0; i < typeOfT.NumField(); i++ {
			val := resultType.Field(i).Interface()
			switch v := val.(type) {
			case int64:
				switch typeOfT.Field(i).Tag.Get("aggregate") {
				case "sum":
					lt.sumInt(i, v)
				case "avg":
					lt.avgInt(i, v)
				}
			case float64:
				switch typeOfT.Field(i).Tag.Get("aggregate") {
				case "sum":
					lt.sumFloat(i, v)
				case "avg":
					lt.avgFloat(i, v)
				}
			}
		}
	}
}

func (lt *LoadTest) sumInt(fieldNum int, val interface{}) {
	thisField := reflect.ValueOf(&lt.AggregatedResults).Elem().Field(fieldNum)
	if valToAdd, ok := val.(int64); ok {
		if currentVal, ok := thisField.Interface().(int64); ok {
			thisField.SetInt(currentVal + valToAdd)
		}
	} else {
		panic("Could not convert struct field to float64")
	}
}

func (lt *LoadTest) avgInt(fieldNum int, val interface{}) {
	thisField := reflect.ValueOf(&lt.AggregatedResults).Elem().Field(fieldNum)
	if valToAdd, ok := val.(int64); ok {
		if currentVal, ok := thisField.Interface().(int64); ok {
			if currentVal == 0 {
				thisField.SetInt(currentVal + valToAdd)
			} else {
				thisField.SetInt((currentVal + valToAdd) / 2)
			}
		}
	} else {
		panic("Could not convert struct field to float64")
	}
}

func (lt *LoadTest) sumFloat(fieldNum int, val interface{}) {
	thisField := reflect.ValueOf(&lt.AggregatedResults).Elem().Field(fieldNum)
	if valToAdd, ok := val.(float64); ok {
		if currentVal, ok := thisField.Interface().(float64); ok {
			thisField.SetFloat(currentVal + valToAdd)
		}
	} else {
		panic("Could not convert struct field to float64")
	}
}

func (lt *LoadTest) avgFloat(fieldNum int, val interface{}) {
	thisField := reflect.ValueOf(&lt.AggregatedResults).Elem().Field(fieldNum)
	if valToAdd, ok := val.(float64); ok {
		if currentVal, ok := thisField.Interface().(float64); ok {
			if currentVal == 0 {
				thisField.SetFloat(currentVal + valToAdd)
			} else {
				thisField.SetFloat((currentVal + valToAdd) / 2)
			}
		}
	} else {
		panic("Could not convert struct field to float64")
	}
}

type TestResult struct {
	LatencyMax           float64 `json:"latency_max,string" aggregate:"avg"`
	LatencyMean          float64 `json:"latency_mean,string" aggregate:"avg"`
	LatencyMin           float64 `json:"latency_min,string" aggregate:"avg"`
	LatencyP50           float64 `json:"latency_p50,string" aggregate:"avg"`
	LatencyP90           float64 `json:"latency_p90,string" aggregate:"avg"`
	LatencyP99           float64 `json:"latency_p99,string" aggregate:"avg"`
	LatencyStdev         float64 `json:"latency_stdev,string" aggregate:"avg"`
	RequestsMax          int64   `json:"requests_max,string" aggregate:"sum"`
	RequestsMean         float64 `json:"requests_mean,string" aggregate:"sum"`
	RequestsMin          int64   `json:"requests_min,string" aggregate:"sum"`
	RequestsP50          int64   `json:"requests_p50,string" aggregate:"sum"`
	RequestsP90          int64   `json:"requests_p90,string" aggregate:"sum"`
	RequestsP99          int64   `json:"requests_p99,string" aggregate:"sum"`
	RequestsStdev        float64 `json:"requests_stdev,string" aggregate:"sum"`
	SummaryBytes         int64   `json:"summary_bytes,string" aggregate:"sum"`
	SummaryDuration      int64   `json:"summary_duration,string" aggregate:"avg"`
	SummaryErrorsConnect int64   `json:"summary_errors_connect,string" aggregate:"sum"`
	SummaryErrorsRead    int64   `json:"summary_errors_read,string" aggregate:"sum"`
	SummaryErrorsStatus  int64   `json:"summary_errors_status,string" aggregate:"sum"`
	SummaryErrorsTimeout int64   `json:"summary_errors_timeout,string" aggregate:"sum"`
	SummaryErrorsWrite   int64   `json:"summary_errors_write,string" aggregate:"sum"`
	SummaryRequests      int64   `json:"summary_requests,string" aggregate:"sum"`
}
