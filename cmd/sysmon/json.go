package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func printSnapshotJSON(snap Snapshot) {
	data, err := json.Marshal(snap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "json:", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
