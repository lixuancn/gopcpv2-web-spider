package module

import (
	"net"
	"strconv"
	"gopcpv2-web-spider/errors"
)

type mAddr struct{
	network string
	address string
}

func(maddr *mAddr)Network()string{
	return maddr.network
}

func (maddr *mAddr)String() string {
	return maddr.address
}

func NewAddr(network, ip string, port uint64)(net.Addr, error){
	if network != "http" && network != "https"{
		return nil, errors.NewIllegalParameterError("无效网络协议")
	}
	if parsedIP := net.ParseIP(ip); parsedIP == nil{
		return nil, errors.NewIllegalParameterError("IP无效")
	}
	return &mAddr{
		network: network,
		address: ip + ":" + strconv.Itoa(int(port)),
	}, nil
}