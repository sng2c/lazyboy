package main

import (
	"encoding/json"
	"lazyboy/tmpl"
)

func tick(path string) {
	// 1. SETUP
	// load config from path

	// make sure consumed directory

	// 2. TAKE
	// take while enough or no more data
	// when EOF move into consumed directory
	// or write file pointer

	// 3. MERGE DATA

}

func main() {
	var i interface{}
	json.Unmarshal([]byte(`{"name":"khs"}`), &i)
	got, _ := tmpl.Resolve("{{ref \"/name\"}}", i)

	print(got)
}
