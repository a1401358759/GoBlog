package entity

import (
	"encoding/xml"
	"reflect"
)

type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
	Meta interface{} `json:"meta"`
}

type Detail struct {
	ErrorCode string `xml:"ErrorCode"`
	Message   string `xml:"Message"`
	Method    string `xml:"method"`
	ID        string `xml:"ID"`
}

type Fault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
	FaultActor  string `xml:"faultactor"`
	Detail      Detail `xml:"detail"`
}

type FaultBody struct {
	Fault Fault `xml:"soap:Fault"`
}

type FaultEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	SOAP    string   `xml:"xmlns:soap,attr"`
	XSD     string   `xml:"xmlns:xsd,attr"`
	XSI     string   `xml:"xmlns:xsi,attr"`

	Body FaultBody `xml:"soap:Body"`
}

func NewResponse() *Response {
	return &Response{
		Meta: map[string]interface{}{},
		Data: map[string]interface{}{},
	}
}

func (r *Response) SetMeta(meta interface{}) *Response {
	r.Meta = meta
	return r
}

func (r *Response) SetData(data interface{}) *Response {
	// 处理为空返回的问题
	// 当此处data为空时，列表需要返回空列表，Map需要返回空Map而不是统一的nil，否则前端需要大量的处理
	if reflect.ValueOf(data).IsZero() {
		switch reflect.TypeOf(data).Kind() {
		case reflect.Slice:
			data = []string{}
		case reflect.Map, reflect.Struct:
			data = make(map[string]string, 0)
		}
	}
	r.Data = data
	return r
}

func (r *Response) SetCode(code int) *Response {
	if r.Code == 0 {
		r.Code = code
	}
	return r
}

func (r *Response) SetMsg(msg string) *Response {
	if len(r.Msg) == 0 {
		r.Msg = msg
	}
	return r
}
