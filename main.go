package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"
)

type RequestFormat struct {
	To_sort [][]int
}

type ResponseFormat struct {
	Sorted_arrays [][]int
	Time_ns       string
}

type ConcChannelResult struct {
	Index        int
	Sorted_array []int
}

func sortSequential(array []int) {
	slices.Sort(array)
}

func sortConcurrent(array []int, index int, channel chan<- ConcChannelResult, waitgroup *sync.WaitGroup) {
	slices.Sort(array)
	channel <- ConcChannelResult{Index: index, Sorted_array: array}
	waitgroup.Done()
}

func processSequential(w http.ResponseWriter, r *http.Request) {
	var data RequestFormat
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response ResponseFormat
	response.Sorted_arrays = data.To_sort

	startTime := time.Now()

	for _, array := range response.Sorted_arrays {
		sortSequential(array)
	}

	response.Time_ns = fmt.Sprint(time.Since(startTime).Nanoseconds())

	d, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func processConcurrent(w http.ResponseWriter, r *http.Request) {
	var data RequestFormat
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response ResponseFormat
	response.Sorted_arrays = data.To_sort

	channel := make(chan ConcChannelResult)
	var waitgroup sync.WaitGroup

	startTime := time.Now()

	for index, array := range response.Sorted_arrays {
		waitgroup.Add(1)
		go sortConcurrent(array, index, channel, &waitgroup)
	}

	go func() {
		waitgroup.Wait()
		close(channel)
	}()

	for result := range channel {
		response.Sorted_arrays[result.Index] = result.Sorted_array
	}

	response.Time_ns = fmt.Sprint(time.Since(startTime).Nanoseconds())

	d, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func main() {
	http.HandleFunc("/process-single", processSequential)
	http.HandleFunc("/process-concurrent", processConcurrent)
	http.ListenAndServe(":8000", nil)
}
