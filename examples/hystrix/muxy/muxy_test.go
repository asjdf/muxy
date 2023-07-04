package examples

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func Test_Example100calls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	fmt.Println("Waiting for Muxy..")

	select {
	case <-time.After(2 * time.Second):
	}

	fmt.Println("Running tests")

	host := fmt.Sprintf("http://api/")
	wait := &sync.WaitGroup{}
	const numberOfRequests = 100

	wait.Add(numberOfRequests)
	for i := 0; i < numberOfRequests; i++ {
		go func() {
			defer wait.Done()
			resp, err := http.Get(host)
			checkErr(err, false, t)
			fmt.Println(resp)

			if resp != nil {
				fmt.Println("\nResponse:")
				r := bufio.NewReader(resp.Body)
				_, _ = r.WriteTo(os.Stdout)
				fmt.Println()
			} else {
				fmt.Println("No response body")
			}
			if resp.StatusCode != 200 {
				t.Errorf("Expected 200 response code, but got %d", resp.StatusCode)
				return
			}
		}()
	}
	fmt.Println("Waiting for all requests to finish...")
	wait.Wait()
}

func checkErr(err error, expected bool, t *testing.T) {
	if err != nil && !expected {
		t.Fatalf("Error not expected: %s", err.Error())

	} else if err == nil && expected {
		t.Fatalf("Error expected, but did not get one")
	}
}
