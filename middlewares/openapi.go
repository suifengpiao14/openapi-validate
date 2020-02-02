package middlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/suifengpiao14/openapi-validate/adapters"
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

var docFile = "doc/openapi.json"

// ValidationRequest validate request params
func ValidationRequest() negroni.Handler {
	return negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

		adapterOpenapi := &adapters.Openapi{
			Request: req,
		}
		doc, err := adapterOpenapi.LoadDoc(docFile)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		adapterOpenapi.Doc = doc

		requestValidationInput, err := adapterOpenapi.GetRequestValidationInput()
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
		var err error
		next(testRes, req) // 替换成httptest，方便后面获取返回体
		respBody := testRes.Body.Bytes()
		if testRes.Code != http.StatusOK { // 非200状态数据不验证
			res.WriteHeader(testRes.Code)
			if _, err := res.Write(respBody); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			return

		}

		adapterOpenapi := &adapters.Openapi{
			Request: req,
		}
		adapterOpenapi.Doc, err = adapterOpenapi.LoadDoc(docFile)
		responseValidationInput, err := adapterOpenapi.GetResponseValidationInput()
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err = openapi3filter.ValidateResponse(nil, responseValidationInput); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		res.WriteHeader(testRes.Code)
		defer responseValidationInput.Body.Close()
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
