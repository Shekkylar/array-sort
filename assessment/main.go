package main

import (
	"net/http"
	"sort"
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
	var sorted [][]int
	var done = make(chan bool)

	for _, subArray := range toSort {
		go func(subArray []int) {
			defer func() { done <- true }()
			sort.Ints(subArray)
			sorted = append(sorted, subArray)
		}(subArray)
	}

	for range toSort {
		<-done
	}

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

	if err := router.Run(":8000"); err != nil {
		panic(err)
	}
}
