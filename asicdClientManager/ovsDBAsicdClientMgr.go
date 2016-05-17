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
	"fmt"
)

type OvsDBAsicdClientMgr struct {
	Val int
}

func (asicdClientMgr *OvsDBAsicdClientMgr) CreateIPv4Neighbor(ipAddr, macAddr string, vlanId, ifIdx int32) (int32, error) {
	fmt.Println(ipAddr, macAddr, vlanId, ifIdx, asicdClientMgr.Val)
	return 0, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) UpdateIPv4Neighbor(ipAddr, macAddr string, vlanId, ifIdx int32) (int32, error) {
	fmt.Println(ipAddr, macAddr, vlanId, ifIdx, asicdClientMgr.Val)
	return 0, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) DeleteIPv4Neighbor(ipAddr string) (int32, error) {
	fmt.Println(ipAddr, asicdClientMgr.Val)
	return 0, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) GetBulkIPv4IntfState(curMark, count int) (*IPv4IntfStateGetInfo, error) {
	fmt.Println("IPv4 Intf State", curMark, count, asicdClientMgr.Val)
	return nil, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) GetBulkPort(curMark, count int) (*PortGetInfo, error) {
	fmt.Println("Port Get info", curMark, count, asicdClientMgr.Val)
	return nil, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) GetBulkPortState(curMark, count int) (*PortStateGetInfo, error) {
	fmt.Println("Port State Get info", curMark, count, asicdClientMgr.Val)
	return nil, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) GetBulkVlanState(curMark, count int) (*VlanStateGetInfo, error) {
	fmt.Println("Vlan State Get info", curMark, count, asicdClientMgr.Val)
	return nil, nil
}

func (asicdClientMgr *OvsDBAsicdClientMgr) GetBulkVlan(curMark, count int) (*VlanGetInfo, error) {
	fmt.Println("Vlan Get info", curMark, count, asicdClientMgr.Val)
	return nil, nil
}
