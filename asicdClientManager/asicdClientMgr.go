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

package asicdClientManager

import (
	"utils/logging"
)

type IPv4IntfState struct {
	IntfRef           string
	IfIndex           int32
	IpAddr            string
	OperState         string
	NumUpEvents       int32
	LastUpEventTime   string
	NumDownEvents     int32
	LastDownEventTime string
	L2IntfType        string
	L2IntfId          int32
}

type IPv4IntfStateGetInfo struct {
	StartIdx          int32
	EndIdx            int32
	Count             int32
	More              bool
	IPv4IntfStateList []IPv4IntfState
}

type Port struct {
	PortNum     int32
	Description string
	PhyIntfType string
	AdminState  string
	MacAddr     string
	Speed       int32
	Duplex      string
	Autoneg     string
	MediaType   string
	Mtu         int32
}

type PortGetInfo struct {
	StartIdx int32
	EndIdx   int32
	Count    int32
	More     bool
	PortList []Port
}

type PortState struct {
	PortNum           int32
	IfIndex           int32
	Name              string
	OperState         string
	NumUpEvents       int32
	LastUpEventTime   string
	NumDownEvents     int32
	LastDownEventTime string
	Pvid              int32
	IfInOctets        int64
	IfInUcastPkts     int64
	IfInDiscards      int64
	IfInErrors        int64
	IfInUnknownProtos int64
	IfOutOctets       int64
	IfOutUcastPkts    int64
	IfOutDiscards     int64
	IfOutErrors       int64
	ErrDisableReason  string
}

type PortStateGetInfo struct {
	StartIdx      int32
	EndIdx        int32
	Count         int32
	More          bool
	PortStateList []PortState
}

type Vlan struct {
	VlanId           int32
	IfIndexList      []int32
	UntagIfIndexList []int32
}

type VlanGetInfo struct {
	StartIdx int32
	EndIdx   int32
	Count    int32
	More     bool
	VlanList []Vlan
}

type VlanState struct {
	VlanId    int32
	VlanName  string
	OperState string
	IfIndex   int32
}

type VlanStateGetInfo struct {
	StartIdx      int32
	EndIdx        int32
	Count         int32
	More          bool
	VlanStateList []VlanState
}

type AsicdClientIntf interface {
	CreateIPv4Neighbor(ipAddr string, macAddr string, vlanId int32, ifIdx int32) (rv int32, err error)
	UpdateIPv4Neighbor(ipAddr string, macAddr string, vlanId int32, ifIdx int32) (rv int32, err error)
	DeleteIPv4Neighbor(ipAddr string) (rv int32, err error)

	GetBulkIPv4IntfState(curMark, count int) (*IPv4IntfStateGetInfo, error)

	GetBulkPort(curMark, count int) (*PortGetInfo, error)
	GetBulkPortState(curMark, count int) (*PortStateGetInfo, error)
	GetBulkVlan(curMark, count int) (*VlanGetInfo, error)
	GetBulkVlanState(curMark, count int) (*VlanStateGetInfo, error)
}

func NewAsicdClientInit(plugin string, paramsFile string, logger *logging.Writer) AsicdClientIntf {
	if plugin == "Flexswitch" {
		clientHdl := getAsicdThriftClientHdl(paramsFile, logger)
		if clientHdl == nil {
			logger.Err("Unable Initialize Asicd Client")
			return nil
		}
		return &FSAsicdClientMgr{clientHdl}
	} else if plugin == "OvsDB" {
		return &OvsDBAsicdClientMgr{100}
	}
	return nil
}
