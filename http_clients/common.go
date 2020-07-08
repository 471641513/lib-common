package http_clients

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/opay-org/lib-common/clients"

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
	rspPtr interface{},
	marshalerOpt ...RespMarshaler) (respBytes []byte, err error) {

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
	proxyReq, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return
	}

	proxyReq.Header.Set("content-type", CONTENT_TYPE_JSON)
	proxyReq.Header.Set(clients.HEADER_CALLER, cli.conf.Caller)
	proxyReq.Header.Set(clients.HEADER_TRACE, ctx.LogId())

	xlog.Info("proxyReq.Header=%+v", proxyReq.Header)
	rspBody, err := cli.Client.Do(proxyReq)
	/*
		rspBody, err := cli.Client.Post(
			url,
			CONTENT_TYPE_JSON,
			reader)
	*/
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
		err = cli.decodeRsp(respBytes, rspPtr, marshalerOpt...)
	}
	return
}

func (cli *HttpClient) GetJsonBody(
	ctx *local_context.LocalContext,
	path string,
	rspPtr interface{},
	marshalerOpt ...RespMarshaler) (respBytes []byte, err error) {

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
		err = cli.decodeRsp(respBytes, rspPtr, marshalerOpt...)
	}
	return
}

func (cli *HttpClient) decodeRsp(respBytes []byte, v interface{}, marshalerOpt ...RespMarshaler) (err error) {
	var marshaler RespMarshaler
	if len(marshalerOpt) == 0 {
		marshaler = DefaultMarshaler
	} else {
		marshaler = marshalerOpt[0]
	}
	errStatus := marshaler.Unmarshal(respBytes, v)
	if errStatus != nil {
		err = errStatus.Err()
	}
	return
}
