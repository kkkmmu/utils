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
	"errors"
	"fmt"
	"net"
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
	ipMask = make(net.IP, 16)
	copy(ipMask, ipNet.Mask)
	ipAddrStr := ip.String()
	//ipMaskStr := net.IP(ipMask).String()
	ipPrefix, err = GetNetowrkPrefixFromStrings(ipAddrStr, (net.IP(ipNet.Mask)).String()) //ipMaskStr)

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
	return ip, nil
}
func IsZeros(p net.IP) bool {
	for i := 0; i < len(p); i++ {
		if p[i] != 0 {
			return false
		}
	}
	return true
}
func IsZerosIPString(ipAddr string) (bool, error) {
	ip, err := GetIP(ipAddr)
	if err != nil {
		fmt.Println("invalid IP address")
		return false, errors.New("Invalid IP address")
	}
	if IsIPv4Mask(ip) {
		return IsZeros(ip[12:15]), nil
	} else {
		return IsZeros(ip), nil
	}
	//fmt.Println("ip:", ip, "len(ip):", len(ip), "ip[12:15]:", ip[12:15], " net.IP(ipAddr):", net.IP(ipAddr))
	return IsZeros(ip), nil
}
func IsIPv4Mask(mask net.IP) bool {
	if IsZeros(mask[0:10]) &&
		mask[10] == 0xff &&
		mask[11] == 0xff {
		return true
	}
	return false
}

func GetPrefixLen(networkMask net.IP) (prefixLen int, err error) {
	//fmt.Println("GetPrefixLen() for mask:", networkMask)
	mask := net.IPMask(networkMask)
	//fmt.Println("mask:", mask)
	if IsIPv4Mask(net.IP(mask)) {
		prefixLen, _ = mask[12:16].Size()
	} else {
		prefixLen, _ = mask.Size()
	}
	//fmt.Println("prefixLen = ", prefixLen, " err:", err)
	return prefixLen, err
}
func GetNetworkPrefix(destNetIp net.IP, networkMask net.IP) (destNet patriciaDB.Prefix, err error) {
	prefixLen, err := GetPrefixLen(networkMask)
	if err != nil {
		fmt.Println("err when getting prefixLen, err= ", err)
		return destNet, errors.New(fmt.Sprintln("Invalid networkmask ", networkMask))
	}
	var netIp net.IP
	vdestMask := net.IPMask(networkMask)
	if IsIPv4Mask(net.IP(vdestMask)) {
		netIp = destNetIp.Mask(vdestMask[12:16])
	} else {
		netIp = destNetIp.Mask(vdestMask)
	}
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
	maskIP, err := GetIP(mask)
	if err != nil {
		fmt.Println("err in getting mask IP for mask string", mask)
		return addr, err
	}
	prefixLen, err := GetPrefixLen(maskIP)
	if err != nil {
		fmt.Println("err in getting prefix len for mask string", mask)
		return addr, err
	}
	addr = (destNetIpAddr.Mask(net.IPMask(maskIP))).String() + "/" + strconv.Itoa(prefixLen)
	return addr, err
}
func CheckIfInRange(testIPAddr, ipAddr string, lowPrefixLen int, highPrefixLen int) bool {
	//fmt.Println("testIPAddr:", testIPAddr, " ipAddr:", ipAddr, " lowPrefixLen:", lowPrefixLen, " highPrefixLen:", highPrefixLen)
	//testAddr := net.ParseIP(testIPAddr)
	testAddr, _, err := net.ParseCIDR(testIPAddr)
	if err != nil {
		fmt.Println("error parsing address:", testIPAddr)
		return false
	}
	if lowPrefixLen == -1 && highPrefixLen == -1 {
		_, cidrnet, err := net.ParseCIDR(ipAddr)
		if err != nil {
			fmt.Println("Error parsing cidr addr ", ipAddr)
			return false
		}
		if cidrnet.Contains(testAddr) == true {
			//fmt.Println(cidrnet, " contains ip:", testAddr)
			return true
		} else {
			fmt.Println(cidrnet, " does not contain ip:", testAddr)
			return false
		}
	}
	baseAddr, _, err := net.ParseCIDR(ipAddr)
	if err != nil {
		fmt.Println("error parsing address:", ipAddr)
		return false
	}
	for idx := lowPrefixLen; idx <= highPrefixLen; idx++ {
		networkAddr := baseAddr.String() + "/" + strconv.Itoa(idx)
		_, cidrnet, err := net.ParseCIDR(networkAddr)
		if err != nil {
			fmt.Println("Error parsing cidr addr ", networkAddr)
			return false
		}
		if cidrnet.Contains(testAddr) == true {
			//fmt.Println(cidrnet, " contains ip:", testAddr)
			return true
		}
	}
	return false
}

func IsIPv6Addr(ipAddr string) bool {
	ip, _, err := net.ParseCIDR(ipAddr)
	if err != nil {
		ip = net.ParseIP(ipAddr)
		if ip == nil {
			return false
		}
	}
	ip1 := ip.To4()
	if ip1 == nil {
		return true
	}
	ip2 := ip.To16()
	if len(ip1) == len(ip2) {
		return true
	}
	return false
}

func IsIPv4Addr(ipAddr string) bool {
	ip, _, err := net.ParseCIDR(ipAddr)
	if err != nil {
		ip = net.ParseIP(ipAddr)
		if ip == nil {
			return false
		}
	}
	ip1 := ip.To4()
	if ip1 == nil {
		return false
	}
	ip2 := ip.To16()
	if len(ip1) != len(ip2) {
		return true
	}
	return false
}
