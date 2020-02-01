package adapters

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/pkg/errors"
)

//Validator interface
type Validator interface {
	ValidateRequest() (err error)
	ValidateResponse() (err error)
}

//Openapi object
type Openapi struct {
	Doc      []byte
	Request  *http.Request
	Response struct {
		ContentType []string
		Body        []byte
		Status      int64
	}
	NewRequestParams []byte
	NewResponseBody  []byte
}

var bodyDecoders = make(map[string]openapi3filter.BodyDecoder)

func (openapi *Openapi) ValidateRequest(requestValidationInput *openapi3filter.RequestValidationInput) (err error) {
	if err = openapi3filter.ValidateRequest(nil, requestValidationInput); err != nil {
		return
	}
	return
}

func (openapi *Openapi) ValidateResponse(responseValidationInput *openapi3filter.ResponseValidationInput) (err error) {

	body := openapi.Request.Response.Body
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	responseValidationInput.SetBodyBytes(data)
	err = openapi3filter.ValidateResponse(nil, responseValidationInput)
	return

}

//FilterRequestParams filter request params
func (openapi *Openapi) FilterRequestParams(input *openapi3filter.RequestValidationInput) (params []byte, err error) {
	options := input.Options
	if options == nil {
		options = &openapi3filter.Options{}
	}
	route := input.Route
	if route == nil {
		err = fmt.Errorf("invalid route")
		return
	}
	operation := route.Operation
	if operation == nil {
		err = fmt.Errorf(" route missing operation")
		return
	}
	operationParameters := operation.Parameters
	pathItemParameters := route.PathItem.Parameters

	// For each parameter of the PathItem
	for _, parameterRef := range pathItemParameters {
		parameter := parameterRef.Value
		if operationParameters != nil {
			if override := operationParameters.GetByInAndName(parameter.In, parameter.Name); override != nil {
				continue
			}
		}
		err = openapi3filter.ValidateParameter(nil, input, parameter)
		if err != nil {
			return
		}
	}

	// For each parameter of the Operation
	for _, parameter := range operationParameters {
		if err = openapi3filter.ValidateParameter(nil, input, parameter.Value); err != nil {
			return
		}
	}

	// RequestBody
	requestBody := operation.RequestBody
	if requestBody != nil && !options.ExcludeRequestBody {
		if err = openapi3filter.ValidateRequestBody(nil, input, requestBody.Value); err != nil {
			return
		}
	}
	return
}

//GetRequestValidationInput get request validation input
func (openapi *Openapi) GetRequestValidationInput() (requestValidationInput *openapi3filter.RequestValidationInput, err error) {
	router, err := openapi.LoadOpenAPI()
	if err != nil {
		return
	}

	if openapi.Request == nil {
		err = fmt.Errorf("Validator Openapi's Request attribute should not be nil ")
	}

	route, pathParams, err := router.FindRoute(openapi.Request.Method, openapi.Request.URL)
	if err != nil {
		return
	}

	requestValidationInput = &openapi3filter.RequestValidationInput{
		Request:    openapi.Request,
		Route:      route,
		PathParams: pathParams,
		AllParams:  make(map[string]interface{}),
	}
	return
}

//GetResponseValidationInput get response validation input
func (openapi *Openapi) GetResponseValidationInput() (responseValidationInput *openapi3filter.ResponseValidationInput, err error) {
	requestValidationInput, err := openapi.GetRequestValidationInput()
	if err != nil {
		return
	}
	var status = 200
	if openapi.Response.Status != 0 {
		status = int(openapi.Response.Status)
	}
	var respContentType = "application/json"
	responseValidationInput = &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 status,
		Options: &openapi3filter.Options{
			TrimAdditionalProperties: true,
		},
		Header: http.Header{
			"Content-Type": []string{
				respContentType,
			},
		},
	}
	return
}

//LoadOpenAPI Load the OpenAPI document and create the router.
func (openapi *Openapi) LoadOpenAPI() (router *openapi3filter.Router, err error) {
	defer func() {
		if r := recover(); r != nil {
			router = nil
			if e, ok := r.(error); ok {
				err = errors.Wrap(e, "Caught panic while trying to load")
			} else {
				err = fmt.Errorf("Caught panic while trying to load")
			}
		}
	}()

	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true

	var u *url.URL
	u, err = url.Parse(openapi.Request.RequestURI)
	if err != nil {
		return
	}

	swagger, err := loader.LoadSwaggerFromDataWithPath(openapi.Doc, u)
	if err != nil {
		return
	}

	// Create a new router using the OpenAPI document's declared paths.
	router = openapi3filter.NewRouter().WithSwagger(swagger)

	return
}
