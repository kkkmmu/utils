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

package asicdClient

import (
	"utils/asicdClient/flexswitch"
	"utils/asicdClient/ovs"
	"utils/commonDefs"
	"utils/logging"
)

type AsicdClientStruct struct {
	Logger *logging.Writer
	NHdl   commonDefs.AsicdNotificationHdl
	NList  commonDefs.AsicdNotification
}

type AsicdClientIntf interface {
	CreateIPv4Neighbor(ipAddr string, macAddr string, vlanId int32, ifIdx int32) (rv int32, err error)
	UpdateIPv4Neighbor(ipAddr string, macAddr string, vlanId int32, ifIdx int32) (rv int32, err error)
	DeleteIPv4Neighbor(ipAddr string) (rv int32, err error)

	GetBulkIPv4IntfState(curMark, count int) (*commonDefs.IPv4IntfStateGetInfo, error)

	GetBulkPort(curMark, count int) (*commonDefs.PortGetInfo, error)
	GetBulkPortState(curMark, count int) (*commonDefs.PortStateGetInfo, error)
	GetBulkVlan(curMark, count int) (*commonDefs.VlanGetInfo, error)
	GetBulkVlanState(curMark, count int) (*commonDefs.VlanStateGetInfo, error)
}

func NewAsicdClientInit(plugin string, paramsFile string, asicdHdl commonDefs.AsicdClientStruct) AsicdClientIntf {
	if plugin == "Flexswitch" {
		clientHdl := flexswitch.GetAsicdThriftClientHdl(paramsFile, asicdHdl.Logger)
		if clientHdl == nil {
			asicdHdl.Logger.Err("Unable Initialize Asicd Client")
			return nil
		}
		flexswitch.InitFSAsicdSubscriber(asicdHdl)
		return &flexswitch.FSAsicdClientMgr{clientHdl}
	} else if plugin == "OvsDB" {
		ovs.InitOvsAsicdSubscriber(asicdHdl)
		return &ovs.OvsAsicdClientMgr{100}
	}
	return nil
}
