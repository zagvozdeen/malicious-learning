package main

import "github.com/zagvozdeen/malicious-learning"

func main() {
	if err := malicious_learning.ParseQuestionsToMarkdown(); err != nil {
		panic(err)
	}
}
