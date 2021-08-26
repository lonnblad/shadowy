package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/lonnblad/shadowy/connectors/shadowerse"
	"github.com/r3labs/diff"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile | log.Lmicroseconds)
}

func main() {
	port := os.Getenv("port")
	secondaryPort := os.Getenv("secondaryPort")
	candidatePort := os.Getenv("candidatePort")

	http.HandleFunc("/shade_requests", shadeRequestsHandler(secondaryPort, candidatePort))
	http.HandleFunc("/shade_responses", shadeResponsesHandler())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func shadeRequestsHandler(secondaryPort, candidatePort string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(201)

		shadeReq, err := shadowerse.DecodeShadeRequest(req)
		if err != nil {
			log.Print(err)
		}

		shadowyID := shadeReq.Header.Get("x-shadowy-id")

		url := fmt.Sprintf("http://127.0.0.1:%s%s", secondaryPort, shadeReq.URL.String())
		secondaryShadeResp, err := proxy(shadeReq, url)
		if err != nil {
			log.Print(err)
		}

		saveShadeResponse(shadowyID, "secondary", secondaryShadeResp)

		url = fmt.Sprintf("http://127.0.0.1:%s%s", candidatePort, shadeReq.URL.String())
		candidateShadeResp, err := proxy(shadeReq, url)
		if err != nil {
			log.Print(err)
		}

		saveShadeResponse(shadowyID, "candidate", candidateShadeResp)
	}
}

func shadeResponsesHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(201)

		shadeResp, err := shadowerse.DecodeShadeResponse(req)
		if err != nil {
			log.Print(err)
		}

		shadowyID := shadeResp.Header.Get("x-shadowy-id")
		saveShadeResponse(shadowyID, "primary", shadeResp)
	}
}

func proxy(shadeReq shadowerse.ShadeRequest, url string) (shadeResp shadowerse.ShadeResponse, err error) {
	buf := bytes.NewBuffer(shadeReq.Body)

	client := new(http.Client)
	httpReq, err := http.NewRequest(shadeReq.Method, url, buf)
	if err != nil {
		err = fmt.Errorf("couldn't create a new request: %w", err)
		return
	}

	for key, value := range shadeReq.Header {
		for _, val := range value {
			httpReq.Header.Add(key, val)
		}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("couldn't execute the request: %w", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("couldn't read body: %w", err)
		return
	}

	shadeResp = shadowerse.CreateShadeResponse(resp, body)
	return
}

var responses = map[string]map[string]shadowerse.ShadeResponse{}

func saveShadeResponse(shadowyID, mode string, shadeResp shadowerse.ShadeResponse) {
	if _, exists := responses[shadowyID]; !exists {
		responses[shadowyID] = map[string]shadowerse.ShadeResponse{}
	}

	responses[shadowyID][mode] = shadeResp
	if len(responses[shadowyID]) != 3 {
		return
	}

	errs := compare(
		responses[shadowyID]["primary"],
		responses[shadowyID]["secondary"],
		responses[shadowyID]["candidate"],
	)

	if len(errs) != 0 {
		log.Print(errs)
	} else {
		log.Print("PASS")
	}
}

func compare(primary, secondary, candidate shadowerse.ShadeResponse) (errs []error) {
	statusCodeErr := compareInterfaces(primary.StatusCode, secondary.StatusCode, candidate.StatusCode)
	if statusCodeErr != nil {
		errs = append(errs, fmt.Errorf("status code error: %w", statusCodeErr))
	}

	headerErr := compareInterfaces(primary.Header, secondary.Header, candidate.Header)
	if headerErr != nil {
		errs = append(errs, fmt.Errorf("header error: %w", headerErr))
	}

	var bodyErr error
	switch primary.Header.Get("content-type") {
	case "application/json":
		bodyErr = compareJSON(primary.Body, secondary.Body, candidate.Body)
	default:
		bodyErr = compareInterfaces(primary.Body, secondary.Body, candidate.Body)
	}

	if bodyErr != nil {
		errs = append(errs, fmt.Errorf("body error: %w", bodyErr))
	}

	return
}

func compareJSON(primaryBody, secondaryBody, candidateBody []byte) (err error) {
	var primaryStruct map[string]interface{}
	if err = json.Unmarshal(primaryBody, &primaryStruct); err != nil {
		return
	}

	var secondaryStruct map[string]interface{}
	if err = json.Unmarshal(secondaryBody, &secondaryStruct); err != nil {
		return
	}

	var candidateStruct map[string]interface{}
	if err = json.Unmarshal(candidateBody, &candidateStruct); err != nil {
		return
	}

	return compareInterfaces(primaryStruct, secondaryStruct, candidateStruct)
}

func compareInterfaces(primary, secondary, candidate interface{}) error {
	noise, err := diff.Diff(primary, secondary)
	if err != nil {
		return err
	}

	rawdiff, err := diff.Diff(primary, candidate)
	if err != nil {
		return err
	}

	var issues []diff.Change
	for _, potentialIssue := range rawdiff {
		var ok bool

		for _, noiseElement := range noise {
			if potentialIssue.Type == noiseElement.Type {
				change, err := diff.Diff(potentialIssue.Path, noiseElement.Path)
				if err != nil {
					return err
				}

				if ok = len(change) == 0; ok {
					break
				}
			}
		}

		if !ok {
			issues = append(issues, potentialIssue)
		}
	}

	if len(issues) > 0 {
		bs, _ := json.MarshalIndent(issues, "", "\t")
		fmt.Println(string(bs))

		return fmt.Errorf("found %d issues", len(issues))
	}

	return nil
}
