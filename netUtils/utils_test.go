//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//       Unless required by applicable law or agreed to in writing, software
//       distributed under the License is distributed on an "AS IS" BASIS,
//       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//       See the License for the specific language governing permissions and
//       limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package netUtils

import (
	"fmt"
	"net"
	"testing"
)

func IPAddrStringToU8List(ipAddr string) []uint8 {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return ip
	}
	return ip
}

func TestGetNetworkPrefix(t *testing.T) {
	fmt.Println("****TestGetNetworkPrefix****")
	ip := "10.1.10.1"
	mask := "255.255.255.0"
	prefix, err := GetNetworkPrefix(net.IP(ip), net.IP(mask))
	fmt.Println("prefix:", prefix, " err:", err)
	fmt.Println("****************")
}

func TestGetPrefixLen(t *testing.T) {
	fmt.Println("****TestGetPrefixLen()****")
	ip := "255.255.255.0"

	netIP, err := GetIP(ip)
	if err != nil {
		fmt.Println("netIP invalid")
	}
	prefixLen, err := GetPrefixLen(netIP)
	fmt.Println("prefixLen:", prefixLen, " err:", err, " for ip:", ip)

	ip = "0.0.0.0"
	netIP, err = GetIP(ip)
	if err != nil {
		fmt.Println("netIP invalid")
	}
	prefixLen, err = GetPrefixLen(netIP)
	fmt.Println("prefixLen:", prefixLen, " err:", err, " for ip:", ip)

	ip = "255.255.255.0"
	parsedIP := IPAddrStringToU8List(ip)
	fmt.Println("parsedIP:", parsedIP, " for ip:", ip)
	prefixLen, err = GetPrefixLen(parsedIP)
	fmt.Println("prefixLen:", prefixLen, " err:", err, " for ip:", ip)
	fmt.Println("**************************")
}

func TestGetNetworkPrefixFromCIDR(t *testing.T) {
	fmt.Println("****TestGetNetworkPrefixFromCIDR****")
	ip := "10.1.10.1/24"
	prefix, err := GetNetworkPrefixFromCIDR(ip)
	fmt.Println("prefix:", prefix, " err:", err, " for ip:", ip)
	ip = "10.1.10.0/24"
	prefix, err = GetNetworkPrefixFromCIDR(ip)
	fmt.Println("prefix:", prefix, " err:", err, " for ip:", ip)
	ip = "192.168.11.1/31"
	prefix, err = GetNetworkPrefixFromCIDR(ip)
	fmt.Println("prefix:", prefix, " err:", err, " for ip:", ip)
	ip = "fe80::/64"
	prefix, err = GetNetworkPrefixFromCIDR(ip)
	fmt.Println("prefix:", prefix, " err:", err, " for ip:", ip)
	fmt.Println("****************")
}
