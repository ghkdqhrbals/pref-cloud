package main

type TestDto struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	NumUsers int    `json:"numUsers"`
	NumReqs  int    `json:"numReqs"`
}
