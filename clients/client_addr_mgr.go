package clients

import (
	"github.com/opay-org/lib-common/xlog"

	"github.com/BurntSushi/toml"
)

type ClientAddr struct {
	Addrs []string `toml:"addrs"`
}

type ClientAddrConf struct {
	AddrMap map[string]*ClientAddr `toml:"addr_map"`
}

type ClientAddrMgr struct {
	conf ClientAddrConf
}

var cliAddrMgr *ClientAddrMgr

func InitClientAddrMap(filepath string) (err error) {
	conf := ClientAddrConf{}
	if _, err = toml.DecodeFile(filepath, &conf); err != nil {
		xlog.Error("failed to init client map")
	}
	return
}

func getAddrFromSvrMgr(svrName string) (addr *ClientAddr) {

	if cliAddrMgr == nil {
		return
	}

	if cliAddrMgr.conf.AddrMap == nil {
		return
	}
	addr, _ = cliAddrMgr.conf.AddrMap[svrName]
	return
}

func GetAddrFromSvrMg(svrName string) (addr *ClientAddr) {
	return getAddrFromSvrMgr(svrName)
}
