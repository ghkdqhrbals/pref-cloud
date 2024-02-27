package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sort"
	"sync"
	"time"
)

func sendPOSTRequest(url string, postData PostData, wg *sync.WaitGroup, results chan<- RequestMetrics, client *http.Client) {

	reqBody, err := json.Marshal(postData)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	ttfb := time.Duration(-1 * time.Millisecond)
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			ttfb = time.Since(start)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := client.Do(req)
	end := time.Now()

	if err != nil {
		fmt.Printf("Error sending POST request: %v\n", err)
		return
	}

	metrics := RequestMetrics{
		StartedAt:   start,
		Duration:    end.Sub(start),
		TTFB:        ttfb,
		Status:      resp.StatusCode,
		CompletedAt: end,
	}
	results <- metrics

	defer resp.Body.Close()

}

func startTest(ctx context.Context, numUsers int, numRequestsPerUser int, url string, wg *sync.WaitGroup, results chan RequestMetrics, client *http.Client) {
	fmt.Printf("startTest")
	// multiple users sending requests
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go send(numRequestsPerUser, ctx, url, wg, results, client)
	}

	go func() {
		wg.Wait()      // wait all users to finish
		close(results) // close the channel to signal that all requests are done
	}()
}

func send(numRequestsPerUser int, ctx context.Context, url string, wg *sync.WaitGroup, results chan RequestMetrics, client *http.Client) {
	defer wg.Done()
	for j := 0; j < numRequestsPerUser; j++ {
		select {
		case <-ctx.Done():
			return
		default:
			postData := PostData{
				ID:      int64(1), // Adjust as needed
				Title:   "Example Title",
				Content: "Example Content",
			}
			sendPOSTRequest(url, postData, wg, results, client)
		}
	}
}

// Helper function to calculate time.Duration percentiles
func calculateDurationPercentiles(durations []time.Duration, percentiles []float64) map[string]time.Duration {
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	results := make(map[string]time.Duration)
	for _, percentile := range percentiles {
		index := int((percentile / 100.0) * float64(len(durations)))
		if index > 0 && index <= len(durations) {
			results[fmt.Sprintf("p%.0f", percentile)] = durations[index-1]
		}
	}
	return results
}

func calculateTPSPercentiles(tpsValues []int, percentiles []float64) map[string]float64 {
	sort.Sort(sort.Reverse(sort.IntSlice(tpsValues)))
	results := make(map[string]float64)
	for _, percentile := range percentiles {
		index := int((percentile / 100.0) * float64(len(tpsValues)))
		if index > 0 && index <= len(tpsValues) {
			results[fmt.Sprintf("p%.0f", percentile)] = float64(tpsValues[index-1])
		}
	}
	return results
}

// ProcessResults aggregates and processes metrics from the channel
func ProcessResults(resultsChan <-chan RequestMetrics, totalUsers int) Results {
	var mttfbs []time.Duration
	statusCodeCount := make(map[int]int)
	tps := make(map[int]int)
	startTime := time.Now()
	var tpsValues []int

	totalRequests := 0
	totalErrors := 0

	for metrics := range resultsChan {
		second := int(metrics.CompletedAt.Sub(startTime).Seconds())
		tps[second]++
		totalRequests++
		statusCodeCount[metrics.Status]++
		if metrics.Status < 200 || metrics.Status >= 300 {
			totalErrors++
		}
		mttfbs = append(mttfbs, metrics.TTFB)
	}
	fmt.Printf("tps %v\n", tps)

	for _, count := range tps {
		tpsValues = append(tpsValues, count)
	}

	totalDuration := time.Since(startTime)

	totalSuccess := totalRequests - totalErrors
	mttfbAverage := time.Duration(0)
	for _, ttfb := range mttfbs {
		mttfbAverage += ttfb
	}
	mttfbAverage /= time.Duration(len(mttfbs))

	mttfbPercentiles := calculateDurationPercentiles(mttfbs, []float64{50, 75, 90, 95, 99})

	// TPS percentiles calculation would depend on collecting TPS at different intervals, not covered here
	tpsPercentiles := calculateTPSPercentiles(tpsValues, []float64{50, 75, 90, 95, 99})
	return Results{
		TotalRequests:    totalRequests,
		TotalErrors:      totalErrors,
		TotalSuccess:     totalSuccess,
		StatusCodeCount:  statusCodeCount,
		TotalUsers:       totalUsers,
		TotalDuration:    totalDuration,
		MTTFBAverage:     mttfbAverage,
		MTTFBPercentiles: mttfbPercentiles,
		TPSAverage:       float64(totalRequests) / totalDuration.Seconds(),
		// Assuming TPSPercentiles would be calculated similarly to MTTFBPercentiles
		TPSPercentiles: tpsPercentiles, // Placeholder, needs interval data
	}
}
