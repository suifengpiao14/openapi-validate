package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/suifengpiao14/openapi-validate/adapters"
	"github.com/suifengpiao14/openapi-validate/config"
)

//Validate controller interface
type Validate interface {
	Request(w http.ResponseWriter, req *http.Request)
	Response(w http.ResponseWriter, req *http.Request)
}

//validate controller
type validate struct {
}

type JsonRequest struct {
	Method      string `json:"method"`
	URL         string `json:"url"`
	ContentType string `json:"contentType"`
	Body        string `json:"body,omitempty"`
}

type JsonResponse struct {
	ContentType string `json:"contentType"`
	Body        string `json:"body,omitempty"`
}

type JsonHTTP struct {
	Request  *JsonRequest  `json:"request"`
	Response *JsonResponse `json:"response,omitempty"`
}

//NewValidate new validate contoller
func NewValidate() Validate {
	return &validate{}
}

//Request openapi request
func (validate *validate) Request(w http.ResponseWriter, req *http.Request) {
	newReq, err := validate.constructNewRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("%s%s", config.ServerConfig.DocPath, "/openapi.json")
	doc, err := ioutil.ReadFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adapterOpenapi := &adapters.Openapi{
		Doc:     doc,
		Request: newReq,
	}
	requestValidationInput, err := adapterOpenapi.GetRequestValidationInput()
	if err != nil {
		return
	}
	if err := adapterOpenapi.ValidateRequest(requestValidationInput); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params, err := json.Marshal(requestValidationInput.AllParams)
	if err != nil {
		return
	}
	w.Write(params)

	return

}

//Response validate response
func (validate *validate) Response(w http.ResponseWriter, req *http.Request) {
	newReq, err := validate.constructNewRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if newReq.Response == nil {
		err := fmt.Errorf("response must be set ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filename := fmt.Sprintf("%s%s", config.ServerConfig.DocPath, "/openapi.json")
	doc, err := ioutil.ReadFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adapterOpenapi := &adapters.Openapi{
		Doc:     doc,
		Request: newReq,
	}
	responseValidationInput, err := adapterOpenapi.GetResponseValidationInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = adapterOpenapi.ValidateResponse(responseValidationInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data []byte
	data, err = ioutil.ReadAll(responseValidationInput.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer newReq.Response.Body.Close()

	key := "Content-Type"
	headerMap := newReq.Response.Header
	if headerMap != nil {
		if value := newReq.Response.Header.Get(key); value != "" {
			w.Header().Add(key, key)
		}
	}

	_, err = w.Write(data)
	return
}

func (validate *validate) constructNewRequest(req *http.Request) (newReq *http.Request, err error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	jsonHTTP := &JsonHTTP{}
	if err = json.Unmarshal(body, jsonHTTP); err != nil {
		return
	}
	if jsonHTTP.Request == nil {
		err = fmt.Errorf("bad request body")
		return
	}
	jsonHTTP.Request.Method = strings.ToUpper(jsonHTTP.Request.Method)
	requestBody := strings.NewReader(jsonHTTP.Request.Body)
	newReq, err = http.NewRequest(jsonHTTP.Request.Method, jsonHTTP.Request.URL, requestBody)
	if err != nil {
		return
	}
	if jsonHTTP.Response == nil || jsonHTTP.Response.Body == "" || jsonHTTP.Response.ContentType == "" {
		return
	}
	responseBody := ioutil.NopCloser(bytes.NewReader([]byte(jsonHTTP.Response.Body)))

	response := &http.Response{
		Request:       newReq,
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		ContentLength: int64(len(jsonHTTP.Response.Body)),
		Header: http.Header{
			"Content-Type": []string{jsonHTTP.Response.ContentType},
		},
		Body: responseBody,
	}
	newReq.Response = response

	return
}
