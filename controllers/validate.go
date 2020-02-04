package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/suifengpiao14/openapi-validate/adapters"
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
	Doc      string        `json:"doc"`
	Request  *JsonRequest  `json:"request"`
	Response *JsonResponse `json:"response,omitempty"`
}

//HTTPResponse response format
//type HTTPResponse struct {
//	Status  int64 `json:"status"`
//	Message string
//	Data    interface{}
//}

//NewValidate new validate contoller
func NewValidate() Validate {
	return &validate{}
}

//Request openapi request
func (validate *validate) Request(w http.ResponseWriter, req *http.Request) {
	jsonHTTP, err := validate.getRequestBody(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newReq, err := validate.constructNewRequest(jsonHTTP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adapterOpenapi := &adapters.Openapi{
		Request: newReq,
	}
	adapterOpenapi.Doc, err = adapterOpenapi.LoadDoc(jsonHTTP.Doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestValidationInput, err := adapterOpenapi.GetRequestValidationInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := openapi3filter.ValidateRequest(nil, requestValidationInput); err != nil {
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
	jsonHTTP, err := validate.getRequestBody(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if jsonHTTP.Response == nil {
		err := fmt.Errorf("response must be set ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newReq, err := validate.constructNewRequest(jsonHTTP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adapterOpenapi := &adapters.Openapi{
		Request: newReq,
	}
	doc, err := adapterOpenapi.LoadDoc(jsonHTTP.Doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	adapterOpenapi.Doc = doc
	responseValidationInput, err := adapterOpenapi.GetResponseValidationInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data []byte
	defer adapterOpenapi.Request.Response.Body.Close()
	data, err = ioutil.ReadAll(adapterOpenapi.Request.Response.Body)
	if err != nil {
		return
	}
	responseValidationInput.SetBodyBytes(data)

	if err := openapi3filter.ValidateResponse(nil, responseValidationInput); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err = ioutil.ReadAll(responseValidationInput.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer newReq.Response.Body.Close()

	if newReq.Response.Header != nil {
		for key := range newReq.Response.Header {
			value := newReq.Response.Header.Get(key)
			w.Header().Add(key, value)
		}
	}

	//	httpResponse := &HTTPResponse{
	//		Status:  http.StatusOK,
	//		Message: "ok",
	//		Data:    data,
	//	}
	//	output, err := json.Marshal(httpResponse)
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (validate *validate) getRequestBody(req *http.Request) (jsonHTTP *JsonHTTP, err error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	jsonHTTP = &JsonHTTP{}
	if err = json.Unmarshal(body, jsonHTTP); err != nil {
		return
	}

	if jsonHTTP.Doc == "" || jsonHTTP.Request == nil {
		err = fmt.Errorf("bad request body")
		return
	}
	jsonHTTP.Request.Method = strings.ToUpper(jsonHTTP.Request.Method)
	return
}

func (validate *validate) constructNewRequest(jsonHTTP *JsonHTTP) (newReq *http.Request, err error) {
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
