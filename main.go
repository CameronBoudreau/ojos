package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const CAPTURAMA_ADDRESS = "http://localhost:8080/capture"
const DEFAULT_IMAGE = "1x1.png"

type Job struct {
	URL               string
	Selectors         []string
	Success           bool
	CapturamaHttpCode int
}

func main() {
	//Create router for API
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", ojosHandler).Methods("POST")

	//Listen on port for requests
	fmt.Println("Listening on port 2190...")
	log.Fatal(http.ListenAndServe(":2190", router))
}

func ojosHandler(w http.ResponseWriter, r *http.Request) {
	job := Job{}
	//Get data from request
	if err := r.ParseForm(); err != nil {
		fmt.Printf("ParseForm() err: %v\n", err)
		InternalServerErrorWriter(w)
		return
	}
	reqURL := r.FormValue("url")
	selector := r.FormValue("selector")
	fmt.Printf("url: %s\nselector: %s\n", reqURL, selector)
	job.URL = reqURL
	for _, val := range strings.Fields(selector) {
		job.Selectors = append(job.Selectors, val)
	}

	//Build params into querystring for service
	formattedURL, err := formatRequestUrl(reqURL, selector)
	if err != nil {
		fmt.Printf("Error formatting requested url and selector for: %v\n", err)
		InternalServerErrorWriter(w)
		return
	}

	//Get response from Capturama service
	resp, err := http.Get(formattedURL)
	if err != nil {
		fmt.Printf("Error contacting capturama service: %v\n", err)
	}

	bytesWritten := 0
	job.CapturamaHttpCode = resp.StatusCode

	//Check status
	fmt.Printf("Status Code from capturama service: %v\n", resp.StatusCode)
	switch job.CapturamaHttpCode {
	case 200, 206:
		//Set headers before write!
		w.WriteHeader(http.StatusOK)

		//Set image to response
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		w.Write(bodyBytes)
		bytesWritten = len(bodyBytes)
		job.Success = true

	default:
		//Set 1x1 pixel as response image
		bytesWritten = writeDefaultImage(w)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(bytesWritten))
	fmt.Printf("\nJob: %+v\n", job)

	//record stats
	analyzeJob(job)
}

func formatRequestUrl(reqURL, selector string) (formattedURL string, err error) {
	fmt.Println("Starting format of URL for Capturama service...")
	//Properly encode selector string
	selector = URLEncodeString(selector)
	// fmt.Printf("Selector after encode: %q\n", selector)

	var Url *url.URL
	Url, err = url.Parse(CAPTURAMA_ADDRESS)
	if err != nil {
		fmt.Printf("Could not parse address:%s\n", CAPTURAMA_ADDRESS)
		return
	}

	parameters := url.Values{}
	parameters.Add("url", reqURL)
	Url.RawQuery = parameters.Encode()
	//url and selector need different encodings to match requirements
	formattedURL = Url.String() + "&dynamic_size_selector=" + selector
	fmt.Printf("Formatted Addess: %q\n", formattedURL)
	return
}

func URLEncodeString(str string) (newString string) {
	t := &url.URL{Path: str}
	newString = t.String()
	return
}

func writeDefaultImage(w http.ResponseWriter) (size int) {
	//Open temp png file from convert
	file, err := os.Open(DEFAULT_IMAGE)
	if err != nil {
		//500
		InternalServerErrorWriter(w)
		fmt.Printf("Error reading temp png file: %q", err)
		return
	}
	defer file.Close()

	//Decode
	img, err := png.Decode(file)
	if err != nil {
		//500
		InternalServerErrorWriter(w)
		fmt.Printf("Error decoding file to png: %q", err)
		return
	}

	//Place in buffer for writing to responseWriter
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		//500
		InternalServerErrorWriter(w)
		fmt.Printf("Error decoding file to png: %q\n", err)
		return
	}

	//Write from buffer
	if _, err := w.Write(buffer.Bytes()); err != nil {
		fmt.Println("Unable to write image.")
	}

	size = buffer.Len()
	return
}

//Writes an internal server error to the response writer
func InternalServerErrorWriter(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}
