package plugin

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type HttpClientConfig struct {
	BaseUrl string `form:"base_url" bson:"base_url" json:"base_url"`
}

type HttpClient struct {
	Instance *resty.Client

	Config HttpClientConfig
}

type HTTPQUERY map[string]string

type HTTPRESPONSE struct {
	Code int         `form:"code" bson:"code" json:"code"`
	Msg  string      `form:"msg" bson:"msg" json:"msg"`
	Data interface{} `form:"data" bson:"data" json:"data"`
}

func NewHttpClient(config HttpClientConfig) (http_client HttpClient) {

	http_client.Config = config
	http_client.Instance = resty.New()

	return
}

func (http_client *HttpClient) GET(url string, query HTTPQUERY) (response HTTPRESPONSE, err error) {

	var (
		resp *resty.Response
	)

	if resp, err = http_client.Instance.R().SetQueryParams(query).Get(http_client.Config.BaseUrl + "/" + url); err != nil {
		return
	}

	if resp.StatusCode() != REQUEST_SUCCESS {
		response.Code = resp.StatusCode()
		return
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		response.Code = REQUEST_RESPONSE_ERR
		response.Data = string(resp.Body())
	}

	return
}

func (http_client *HttpClient) POST(url string, query HTTPQUERY, body interface{}) (response HTTPRESPONSE, err error) {

	var (
		resp *resty.Response
	)

	if resp, err = http_client.Instance.R().SetQueryParams(query).SetBody(body).Post(http_client.Config.BaseUrl + "/" + url); err != nil {
		return
	}

	if resp.StatusCode() != REQUEST_SUCCESS {
		response.Code = resp.StatusCode()
		return
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		response.Code = REQUEST_RESPONSE_ERR
		response.Data = string(resp.Body())
	}

	return
}

func (http_client *HttpClient) DELETE(url string, query HTTPQUERY) (response HTTPRESPONSE, err error) {

	var (
		resp *resty.Response
	)

	if resp, err = http_client.Instance.R().SetQueryParams(query).Delete(http_client.Config.BaseUrl + "/" + url); err != nil {
		return
	}

	if resp.StatusCode() != REQUEST_SUCCESS {
		response.Code = resp.StatusCode()
		return
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		response.Code = REQUEST_RESPONSE_ERR
		response.Data = string(resp.Body())
	}

	return
}
