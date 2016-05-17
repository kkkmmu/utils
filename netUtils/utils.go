//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __  
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  | 
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  | 
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   | 
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  | 
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__| 
//                                                                                                           

// netUtils.go
package netUtils

import (
	"net"
	"errors"
	"fmt"
	"strconv"
	"utils/patriciaDB"
)
func GetNetowrkPrefixFromStrings(ipAddr string, mask string) (prefix patriciaDB.Prefix, err error) {
	destNetIpAddr, err := GetIP(ipAddr)
	if err != nil {
		fmt.Println("destNetIpAddr invalid")
		return prefix, err
	}
	networkMaskAddr, err := GetIP(mask)
	if err != nil {
		fmt.Println("networkMaskAddr invalid")
		return prefix, err
	}
	prefix, err = GetNetworkPrefix(destNetIpAddr, networkMaskAddr)
	if err != nil {
		fmt.Println("err=", err)
		return prefix, err
	}
	return prefix, err
}
func GetNetworkPrefixFromCIDR(ipAddr string) (ipPrefix patriciaDB.Prefix, err error) {
	var ipMask net.IP
	ip, ipNet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return ipPrefix, err
	}
	ipMask = make(net.IP, 4)
	copy(ipMask, ipNet.Mask)
	ipAddrStr := ip.String()
	ipMaskStr := net.IP(ipMask).String()
	ipPrefix ,err= GetNetowrkPrefixFromStrings(ipAddrStr, ipMaskStr)
    return ipPrefix, err
}
func GetIPInt(ip net.IP) (ipInt int, err error) {
	if ip == nil {
		fmt.Printf("ip address %v invalid\n", ip)
		return ipInt, errors.New("Invalid destination network IP Address")
	}
	ip = ip.To4()
	parsedPrefixIP := int(ip[3]) | int(ip[2])<<8 | int(ip[1])<<16 | int(ip[0])<<24
	ipInt = parsedPrefixIP
	return ipInt, nil
}

func GetIP(ipAddr string) (ip net.IP, err error) {
	ip = net.ParseIP(ipAddr)
	if ip == nil {
		return ip, errors.New("Invalid destination network IP Address")
	}
	ip = ip.To4()
	return ip, nil
}

func GetPrefixLen(networkMask net.IP) (prefixLen int, err error) {
	ipInt, err := GetIPInt(networkMask)
	if err != nil {
		return -1, err
	}
	for prefixLen = 0; ipInt != 0; ipInt >>= 1 {
		prefixLen += ipInt & 1
	}
	return prefixLen, nil
}

func GetNetworkPrefix(destNetIp net.IP, networkMask net.IP) (destNet patriciaDB.Prefix, err error) {
	prefixLen, err := GetPrefixLen(networkMask)
	if err != nil {
		fmt.Println("err when getting prefixLen, err= ", err)
		return destNet, err
	}
	vdestMask := net.IPv4Mask(networkMask[0], networkMask[1], networkMask[2], networkMask[3])
	netIp := destNetIp.Mask(vdestMask)
	numbytes := prefixLen / 8
	if (prefixLen % 8) != 0 {
		numbytes++
	}
	destNet = make([]byte, numbytes)
	for i := 0; i < numbytes; i++ {
		destNet[i] = netIp[i]
	}
	return destNet, err
}
func GetCIDR(ipAddr string, mask string) (addr string, err error) {
	destNetIpAddr, err := GetIP(ipAddr)
	if err != nil {
		fmt.Println("destNetIpAddr invalid")
		return addr, err
	}
	maskIP,err:=GetIP(mask)
	if err != nil {
       fmt.Println("err in getting mask IP for mask string", mask)
	   return addr, err
	}
	prefixLen,err := GetPrefixLen(maskIP)
	if err != nil {
	   fmt.Println("err in getting prefix len for mask string", mask)
	   return addr, err
	}
	addr = (destNetIpAddr.Mask(net.IPMask(maskIP))).String() + "/" + strconv.Itoa(prefixLen)
	return addr, err
}