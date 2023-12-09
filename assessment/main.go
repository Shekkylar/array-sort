package main

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type InputData struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponseData struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNs       int64   `json:"time_ns"`
}

func sortSequential(toSort [][]int) [][]int {
	sorted := make([][]int, len(toSort))
	for i, subArray := range toSort {
		sorted[i] = make([]int, len(subArray))
		copy(sorted[i], subArray)
		sort.Ints(sorted[i])
	}
	return sorted
}

func sortConcurrent(toSort [][]int) [][]int {
	var wg sync.WaitGroup
	var mu sync.Mutex

	sorted := make([][]int, len(toSort))

	for i, subArray := range toSort {
		wg.Add(1)
		go func(index int, arr []int) {
			defer wg.Done()
			sort.Ints(arr)
			mu.Lock()
			sorted[index] = arr
			mu.Unlock()
		}(i, append([]int{}, subArray...)) // Create a copy of the subarray for each goroutine
	}

	wg.Wait()
	return sorted
}

func processSingle(c *gin.Context) {
	var inputData InputData
	if err := c.BindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()
	sortedArrays := sortSequential(inputData.ToSort)
	timeTaken := time.Since(startTime)

	response := ResponseData{
		SortedArrays: sortedArrays,
		TimeNs:       timeTaken.Nanoseconds(),
	}

	c.JSON(http.StatusOK, response)
}

func processConcurrent(c *gin.Context) {
	var inputData InputData
	if err := c.BindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()
	sortedArrays := sortConcurrent(inputData.ToSort)
	timeTaken := time.Since(startTime)

	response := ResponseData{
		SortedArrays: sortedArrays,
		TimeNs:       timeTaken.Nanoseconds(),
	}

	c.JSON(http.StatusOK, response)
}

func main() {
	router := gin.Default()

	router.POST("/process-single", processSingle)
	router.POST("/process-concurrent", processConcurrent)

	if err := router.Run(":10000"); err != nil {
		panic(err)
	}
}
