package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	inputFileName := "input.txt"
	outputFileName := "output.txt"

	// Open the input file for reading
	inputFile, err := os.Open(inputFileName)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Create a scanner to read the input file line by line
	scanner := bufio.NewScanner(inputFile)

	// Custom HTTP client with redirect policy
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // allows capturing the redirect status
		},
	}

	// Read and process each line
	for scanner.Scan() {
		line := scanner.Text()
		// Write both http and https versions to the output file
		fmt.Fprintln(outputFile, "http://"+line)
		fmt.Fprintln(outputFile, "https://"+line)

		// Process each URL for both http and https
		for _, prefix := range []string{"http://", "https://"} {
			url := prefix + line
			response, err := client.Get(url)
			if err != nil {
				fmt.Println("Error fetching URL:", url, err)
				continue
			}
			defer response.Body.Close()

			// Follow redirect if the status code indicates a redirection
			if response.StatusCode >= 300 && response.StatusCode <= 399 {
				location, err := response.Location()
				if err != nil {
					fmt.Println("Error getting redirect location for URL:", url, err)
					continue
				}
				fmt.Println("Following redirect from", url, "to", location.String())
				url = location.String()
				response, err = client.Get(url)
				if err != nil {
					fmt.Println("Error fetching redirect URL:", url, err)
					continue
				}
				defer response.Body.Close()
			}

			// Create directory specific to the URL
			safeURL := strings.ReplaceAll(url, "/", "_")
			safeURL = strings.ReplaceAll(safeURL, ":", "_")
			safeURL = strings.ReplaceAll(safeURL, ".", "_")
			dirName := safeURL
			if err := os.Mkdir(dirName, 0755); err != nil {
				fmt.Println("Error creating directory:", dirName, err)
				continue
			}

			// Save the response body
			bodyFile, err := os.Create(fmt.Sprintf("%s/body.html", dirName))
			if err != nil {
				fmt.Println("Error creating body file for URL:", url, err)
				continue
			}
			defer bodyFile.Close()

			_, err = io.Copy(bodyFile, response.Body)
			if err != nil {
				fmt.Println("Error saving body for URL:", url, err)
				continue
			}

			// Save the headers
			headerFile, err := os.Create(fmt.Sprintf("%s/headers.txt", dirName))
			if err != nil {
				fmt.Println("Error creating header file for URL:", url, err)
				continue
			}
			defer headerFile.Close()

			for key, values := range response.Header {
				fmt.Fprintln(headerFile, key+":", strings.Join(values, ","))
			}
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
	}
}
