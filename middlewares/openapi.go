package middlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/suifengpiao14/openapi-validate/config"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

//Pagination 分页器
type Pagination struct {
	Size  int `json:"pize"`
	Page  int `json:"page"`
	Total int `json:"total"`
}

//ResponseBean 返回体
type ResponseBean struct {
	Msg        string                 `json:"msg"`
	Code       int                    `json:"code"`
	Record     interface{}            `json:"record"`
	List       interface{}            `json:"list"`
	Params     map[string]interface{} `json:"params"`
	Pagination *Pagination            `json:"pagination"`
}

// ValidationRequest validate request params
func ValidationRequest() negroni.Handler {
	return negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

		requestValidationInput, err := GetRequestValidationInput(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := openapi3filter.ValidateRequest(nil, requestValidationInput); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		next(res, req)
		return
	})
}

// ValidateResponse validate response body
func ValidateResponse() negroni.Handler {
	return negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		testRes := httptest.NewRecorder()
		next(testRes, req) // 替换成httptest，方便后面获取返回体
		respBody := testRes.Body.Bytes()
		if testRes.Code != http.StatusOK { // 验证正常返回数据
			res.WriteHeader(testRes.Code)
			if _, err := res.Write(respBody); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			return

		}

		requestValidationInput, err := GetRequestValidationInput(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		var (
			respStatus      = http.StatusOK
			respContentType = "application/json"
		)

		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 respStatus,
			Header: http.Header{
				"Content-Type": []string{
					respContentType,
				},
			},
		}
		responseValidationInput.SetBodyBytes(respBody)
		err = openapi3filter.ValidateResponse(nil, responseValidationInput)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.WriteHeader(testRes.Code)
		// Read all
		newRespBody, err := ioutil.ReadAll(responseValidationInput.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := res.Write(newRespBody); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}

		return
	})
}

func LoadDoc(req *http.Request) (doc []byte, err error) {
	filename := fmt.Sprintf("%s%s", config.ServerConfig.DocPath, "/openapi.json")
	doc, err = ioutil.ReadFile(filename)
	return
}

//GetRequestValidationInput get request validation input
func GetRequestValidationInput(req *http.Request) (requestValidationInput *openapi3filter.RequestValidationInput, err error) {

	doc, err := LoadDoc(req)
	if err != nil {
		return
	}
	_, router, err := LoadOpenAPI(req.RequestURI, doc)

	if err != nil {
		return
	}

	route, pathParams, err := router.FindRoute(req.Method, req.URL)
	if err != nil {
		return
	}

	requestValidationInput = &openapi3filter.RequestValidationInput{
		Request:    req,
		Route:      route,
		PathParams: pathParams,
	}
	return
}

// Load the OpenAPI document and create the router.
func LoadOpenAPI(uri string, data []byte) (swagger *openapi3.Swagger, router *openapi3filter.Router, err error) {
	defer func() {
		if r := recover(); r != nil {
			swagger = nil
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
	u, err = url.Parse(uri)
	if err != nil {
		return
	}

	swagger, err = loader.LoadSwaggerFromDataWithPath(data, u)
	if err != nil {
		return
	}

	// Create a new router using the OpenAPI document's declared paths.
	router = openapi3filter.NewRouter().WithSwagger(swagger)

	return
}

//GetMessageFromRequestError 从json schema 验证错误中提取错误信息
func GetMessageFromRequestError(err error) (*ResponseBean, bool) {
	output := &ResponseBean{}
	if requestError, ok := err.(*openapi3filter.RequestError); ok {
		if schemaError, ok := requestError.Err.(*openapi3.SchemaError); ok {
			externalDocs := schemaError.Schema.ExternalDocs
			if externalDocsMap, ok := externalDocs.(map[string]interface{}); ok {
				if xErrorInterface, ok := externalDocsMap["x-error"]; ok {
					if xError, ok := xErrorInterface.(map[string]interface{}); ok {
						if codeInterface, ok := xError["code"]; ok {
							if code, ok := codeInterface.(int); ok {
								output.Code = code
							}
						}
						if msgInterface, ok := xError["msg"]; ok {
							if msg, ok := msgInterface.(string); ok {
								output.Msg = fmt.Sprintf("%s:%s", requestError.Parameter.Name, msg)
							}
						}
						return output, true
					}

				}
			}

		}
	}
	return output, false
}
