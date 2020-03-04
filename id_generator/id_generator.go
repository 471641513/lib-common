package id_generator

import (
	"errors"
	"net"
	"time"

	"github.com/sony/sonyflake"
)

func NewIdGenerator() (
	sf *sonyflake.Sonyflake,
	err error) {
	t := time.Date(2019, 11, 11, 11, 11, 11, 11, time.UTC)

	ip, _ := lower16BitPrivateIP()
	settings := sonyflake.Settings{
		StartTime: t,
		MachineID: func() (u uint16, e error) {
			u = ip
			return
		},
	}

	sf = sonyflake.NewSonyflake(settings)

	if sf == nil {
		err = errors.New("sonyflake not created")
		return
	}

	return sf, nil
}

func PrivateIpv4() (net.IP, error) {
	return privateIPv4()
}
func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
