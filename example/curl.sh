#!/usr/env bash


# validate request
curl -XPOST http://127.0.0.1:8000/api/v1/validate/request -H 'Content-Type:application/json' -d '{"doc":"./doc/test-openapi.json","request":{"method":"GET","url":"/api/v1/credit/card","contentType":"application/json"}}'

# validate response 
curl -XPOST http://127.0.0.1:8000/api/v1/validate/response -H 'Content-Type:application/json' -d '{"doc":"./doc/test-openapi.json","request":{"method":"GET","url":"/api/v1/credit/card","contentType":"application/json"},"response":{"contentType":"application/json","body":"{\"name\":\" world\",\"hello\":\"hello\",\"user\":[{\"id\":1,\"name\":\"test\"},{\"id\":2,\"name\":\"test2\",\"more\":1}]}"}}'
# {"name":" world","user":[{"id":1,"name":"test"},{"id":2,"name":"test2"}]}
