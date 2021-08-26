package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/lonnblad/shadowy/connectors/shadowerse"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const modePrimary = "primary"

func main() {
	port := os.Getenv("port")

	h := handler{}
	h.mode = os.Getenv("mode")
	h.servicePort = os.Getenv("service_port")
	h.shadowersePort = os.Getenv("shadowerse_port")

	log.Fatal(http.ListenAndServe(":"+port, &h))
}

type handler struct {
	mode           string
	servicePort    string
	shadowersePort string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	url := "http://127.0.0.1:%s" + req.URL.String()

	shadowyID := uuid.New().String()

	defer req.Body.Close()
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	if h.mode == modePrimary {
		go func(shadowyID string) {
			shadeReq := shadowerse.CreateShadeRequest(req, reqBody)
			shadeReq.Header.Add("x-shadowy-id", shadowyID)
			err := shadowerse.SendShadeRequest(h.shadowersePort, shadeReq)

			if err != nil {
				log.Print(err)
			}
		}(shadowyID)
	}

	resp, err := proxy(req, reqBody, fmt.Sprintf(url, h.servicePort))
	if err != nil {
		log.Fatal(err)
	}

	respBody, err := readResponseData(resp)
	if err != nil {
		log.Fatal(err)
	}

	if h.mode == modePrimary {
		go func(shadowyID string) {
			shadeResp := shadowerse.CreateShadeResponse(resp, respBody)
			shadeResp.Header.Add("x-shadowy-id", shadowyID)
			err := shadowerse.SendShadeResponse(h.shadowersePort, shadeResp)

			if err != nil {
				log.Print(err)
			}
		}(shadowyID)
	}

	for key, value := range resp.Header {
		for _, val := range value {
			w.Header().Add(key, val)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err = bytes.NewBuffer(respBody).WriteTo(w); err != nil {
		log.Fatal(err)
	}
}

func proxy(req *http.Request, body []byte, url string) (resp *http.Response, err error) {
	buf := bytes.NewBuffer(body)

	client := new(http.Client)
	httpReq, err := http.NewRequest(req.Method, url, buf)
	if err != nil {
		err = fmt.Errorf("couldn't create a new request: %w", err)
		return
	}

	for key, value := range req.Header {
		for _, val := range value {
			httpReq.Header.Add(key, val)
		}
	}

	resp, err = client.Do(httpReq)
	if err != nil {
		err = fmt.Errorf("couldn't execute the request: %w", err)
		return
	}

	return
}

func readResponseData(resp *http.Response) (bs []byte, err error) {
	defer resp.Body.Close()

	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("couldn't read body: %w", err)
		return
	}

	return
}
