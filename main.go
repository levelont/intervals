package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type fileIndex struct {
	key  interval
	file *os.File
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "f", "", "path to file containing list of intervals to merge.")
	flag.Parse()

	if filePath != "" {
		res, err := processFile(filePath, fileChunkSizeFromEnv())
		if err != nil {
			log.Fatalf("failed to process file %q: %s\n", filePath, err.Error())
		}

		fmt.Printf("result written to file %q\n", res)
	} else {
		if len(os.Args) != 2 {
			fmt.Println("usage: go run . \"INTERVAL_LIST\"")
			fmt.Println("example: go run . \"[1,2] [2,3]\"")
			os.Exit(1)
		}

		res, err := processString(os.Args[1])
		if err != nil {
			log.Fatalf("failed to process input: %s\n", err.Error())
		}
		fmt.Println(IntervalListToString(res))
	}
}

// fileChunkSizeFromEnv reads FILE_CHUNK_SIZE_MB from
// the environment and returns it.
// If the variable is not set, it returns a default.
func fileChunkSizeFromEnv() int {
	var fileChunkSize int
	fileChunkSizeStr, ok := os.LookupEnv("FILE_CHUNK_SIZE_MB")
	if ok {
		var err error
		fileChunkSize, err = strconv.Atoi(fileChunkSizeStr)
		if err != nil {
			log.Fatalln("FILE_CHUNK_SIZE_MB must be a number greater than zero.")
		}
	} else {
		// Default: 1MB
		fileChunkSize = 1
	}

	return fileChunkSize * 1024 * 1024
}
