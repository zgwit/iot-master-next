package plugin

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
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

func (http_client *HttpClient) PROXY(ctx *gin.Context, url string) {

	response, query, body, err := HTTPRESPONSE{}, HTTPQUERY{}, any(nil), any(nil)

	for key, values := range ctx.Request.URL.Query() {
		for _, value := range values {
			query[key] = value
		}
	}

	if err = ctx.Bind(&body); err != nil {
		HttpFailure(ctx, "请求失败，请稍后重试", REQUEST_SERVER_ERR, err)
		return
	}

	switch ctx.Request.Method {

	case "GET":
		response, err = http_client.GET(url, query)
	case "POST":
		response, err = http_client.POST(url, query, body)
	case "DELETE":
		response, err = http_client.DELETE(url, query)

	default:
		HttpFailure(ctx, "不支持的类型", REQUEST_FAIL, "Proxy仅支持：GET、POST、DELETE")
	}

	if err != nil {
		HttpFailure(ctx, "请求失败，请稍后重试", REQUEST_SERVER_ERR, err)
	}

	HttpDefault(ctx, response.Msg, response.Code, response.Data)
}
