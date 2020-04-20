package http_clients

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/metrics"
	"github.com/opay-org/lib-common/utils"
	"github.com/opay-org/lib-common/xlog"

	jsoniter "github.com/json-iterator/go"
)

const (
	CONTENT_TYPE_JSON = "application/json;charset=UTF-8"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type HttpClientConfig struct {
	Addrs     []string `toml:"addrs"`
	TimeoutMs int      `toml:"timeout_ms"`
	Caller    string   `toml:"caller"`
	Https     bool     `toml:"https"`
}

type HttpClient struct {
	conf HttpClientConfig
	*http.Client

	metrics.MetricsBase
}

func NewHttpClient(conf HttpClientConfig) (cli *HttpClient, err error) {
	cli = &HttpClient{
		conf: conf,
		Client: &http.Client{
			Timeout: time.Duration(conf.TimeoutMs) * time.Millisecond,
		},
	}
	return
}
func (cli *HttpClient) Conf() HttpClientConfig {
	return cli.conf
}
func (cli *HttpClient) GetAddr() string {
	if len(cli.conf.Addrs) == 1 {
		return cli.conf.Addrs[0]
	}
	return cli.conf.Addrs[rand.Intn(len(cli.conf.Addrs))]
}

func (cli *HttpClient) CreateMetrics(
	prefix string,
	buckets []float64,
	timecostLables []string,
	countLables []string) {
	if nil == buckets {
		buckets = []float64{5, 10, 60, 200, 500}
	}
	cli.MetricsBase.CreateMetrics(prefix, buckets, timecostLables)
	cli.MetricsBase.CreateMetricsCountVec(prefix, "grpc", "cnt", countLables)
}

func (cli *HttpClient) PostJsonBody(
	ctx *local_context.LocalContext,
	path string,
	data interface{},
	rspPtr interface{}) (respBytes []byte, err error) {

	t0 := time.Now()
	defer func() {
		timeCost := utils.CalTimecost(t0)
		if err != nil {
			xlog.Error("logid=%v||_http_failed||method=POST||path=%v||time_cost=%v||data=%v||err=%+v",
				ctx.LogId(), path, timeCost, utils.MustString(data), err)
		} else {
			xlog.Debug("logid=%v||_http_succ||method=POST||path=%v||time_cost=%v||data=%v||respBytes=%s",
				ctx.LogId(), path, timeCost, utils.MustString(data), respBytes)
		}
	}()
	bytesData, err := json.Marshal(data)
	if err != nil {
		return
	}

	reader := bytes.NewReader(bytesData)
	host := cli.GetAddr()
	url := fmt.Sprintf("http://%v%v", host, path)
	if cli.conf.Https {
		url = fmt.Sprintf("https://%v%v", host, path)
	}

	rspBody, err := cli.Client.Post(
		url,
		CONTENT_TYPE_JSON,
		reader)
	if err != nil {
		return
	}
	if rspBody == nil {
		xlog.Warn("logid=%v||rsp body nil", ctx.LogId())
		return
	}
	respBytes, err = ioutil.ReadAll(rspBody.Body)
	if err != nil {
		return
	}
	if rspPtr != nil {
		err = json.Unmarshal(respBytes, rspPtr)
	}
	return
}

func (cli *HttpClient) GetJsonBody(
	ctx *local_context.LocalContext,
	path string,
	rspPtr interface{}) (respBytes []byte, err error) {

	t0 := time.Now()
	defer func() {
		timeCost := utils.CalTimecost(t0)
		if err != nil {
			xlog.Error("logid=%v||_http_failed||method=GET||path=%v||time_cost=%v||err=%+v",
				ctx.LogId(), path, timeCost, err)
		} else {
			xlog.Debug("logid=%v||_http_succ||method=GET||path=%v||time_cost=%v||respBytes=%s",
				ctx.LogId(), path, timeCost, respBytes)
		}
	}()

	host := cli.GetAddr()
	url := fmt.Sprintf("http://%v%v", host, path)
	if cli.conf.Https {
		url = fmt.Sprintf("https://%v%v", host, path)
	}

	rspBody, err := cli.Client.Get(url)
	if err != nil {
		return
	}
	if rspBody == nil {
		xlog.Warn("logid=%v||rsp body nil", ctx.LogId())
		return
	}
	respBytes, err = ioutil.ReadAll(rspBody.Body)
	if err != nil {
		return
	}
	if rspPtr != nil {
		err = json.Unmarshal(respBytes, rspPtr)
	}
	return
}
