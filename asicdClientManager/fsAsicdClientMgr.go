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
	"asicdInt"
	"asicdServices"
	"encoding/json"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"io/ioutil"
	"strconv"
	"time"
	"utils/ipcutils"
	"utils/logging"
)

type AsicdClient struct {
	ClientBase
	ClientHdl *asicdServices.ASICDServicesClient
}

type ClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

type ClientBase struct {
	Address            string
	Transport          thrift.TTransport
	PtrProtocolFactory *thrift.TBinaryProtocolFactory
}

type FSAsicdClientMgr struct {
	ClientHdl *asicdServices.ASICDServicesClient
}

func (asicdClientMgr *FSAsicdClientMgr) CreateIPv4Neighbor(ipAddr, macAddr string, vlanId, ifIdx int32) (int32, error) {
	return asicdClientMgr.ClientHdl.CreateIPv4Neighbor(ipAddr, macAddr, vlanId, ifIdx)
}

func (asicdClientMgr *FSAsicdClientMgr) UpdateIPv4Neighbor(ipAddr, macAddr string, vlanId, ifIdx int32) (int32, error) {
	return asicdClientMgr.ClientHdl.UpdateIPv4Neighbor(ipAddr, macAddr, vlanId, ifIdx)
}

func (asicdClientMgr *FSAsicdClientMgr) DeleteIPv4Neighbor(ipAddr string) (int32, error) {
	return asicdClientMgr.ClientHdl.DeleteIPv4Neighbor(ipAddr, "00:00:00:00:00:00", 0, 0)
}

func (asicdClientMgr *FSAsicdClientMgr) GetBulkIPv4IntfState(curMark, count int) (*IPv4IntfStateGetInfo, error) {
	bulkInfo, err := asicdClientMgr.ClientHdl.GetBulkIPv4IntfState(asicdServices.Int(curMark), asicdServices.Int(count))
	if bulkInfo == nil {
		return nil, err
	}
	var ipv4Info IPv4IntfStateGetInfo
	ipv4Info.StartIdx = int32(bulkInfo.StartIdx)
	ipv4Info.EndIdx = int32(bulkInfo.EndIdx)
	ipv4Info.Count = int32(bulkInfo.Count)
	ipv4Info.More = bulkInfo.More
	ipv4Info.IPv4IntfStateList = make([]IPv4IntfState, int(ipv4Info.Count))
	for idx := 0; idx < int(ipv4Info.Count); idx++ {
		ipv4Info.IPv4IntfStateList[idx].IntfRef = bulkInfo.IPv4IntfStateList[idx].IntfRef
		ipv4Info.IPv4IntfStateList[idx].IfIndex = bulkInfo.IPv4IntfStateList[idx].IfIndex
		ipv4Info.IPv4IntfStateList[idx].IpAddr = bulkInfo.IPv4IntfStateList[idx].IpAddr
		ipv4Info.IPv4IntfStateList[idx].OperState = bulkInfo.IPv4IntfStateList[idx].OperState
		ipv4Info.IPv4IntfStateList[idx].NumUpEvents = bulkInfo.IPv4IntfStateList[idx].NumUpEvents
		ipv4Info.IPv4IntfStateList[idx].LastUpEventTime = bulkInfo.IPv4IntfStateList[idx].LastUpEventTime
		ipv4Info.IPv4IntfStateList[idx].NumDownEvents = bulkInfo.IPv4IntfStateList[idx].NumDownEvents
		ipv4Info.IPv4IntfStateList[idx].LastDownEventTime = bulkInfo.IPv4IntfStateList[idx].LastDownEventTime
		ipv4Info.IPv4IntfStateList[idx].L2IntfType = bulkInfo.IPv4IntfStateList[idx].L2IntfType
		ipv4Info.IPv4IntfStateList[idx].L2IntfId = bulkInfo.IPv4IntfStateList[idx].L2IntfId
	}
	return &ipv4Info, nil
}

func (asicdClientMgr *FSAsicdClientMgr) GetBulkPort(curMark, count int) (*PortGetInfo, error) {
	bulkInfo, err := asicdClientMgr.ClientHdl.GetBulkPort(asicdServices.Int(curMark), asicdServices.Int(count))
	if bulkInfo == nil {
		return nil, err
	}
	var portInfo PortGetInfo
	portInfo.StartIdx = int32(bulkInfo.StartIdx)
	portInfo.EndIdx = int32(bulkInfo.EndIdx)
	portInfo.Count = int32(bulkInfo.Count)
	portInfo.More = bulkInfo.More
	portInfo.PortList = make([]Port, int(portInfo.Count))
	for idx := 0; idx < int(portInfo.Count); idx++ {
		portInfo.PortList[idx].PortNum = bulkInfo.PortList[idx].PortNum
		portInfo.PortList[idx].Description = bulkInfo.PortList[idx].Description
		portInfo.PortList[idx].PhyIntfType = bulkInfo.PortList[idx].PhyIntfType
		portInfo.PortList[idx].AdminState = bulkInfo.PortList[idx].AdminState
		portInfo.PortList[idx].MacAddr = bulkInfo.PortList[idx].MacAddr
		portInfo.PortList[idx].Speed = bulkInfo.PortList[idx].Speed
		portInfo.PortList[idx].Duplex = bulkInfo.PortList[idx].Duplex
		portInfo.PortList[idx].Autoneg = bulkInfo.PortList[idx].Autoneg
		portInfo.PortList[idx].MediaType = bulkInfo.PortList[idx].MediaType
		portInfo.PortList[idx].Mtu = bulkInfo.PortList[idx].Mtu
	}
	return &portInfo, nil
}

func (asicdClientMgr *FSAsicdClientMgr) GetBulkPortState(curMark, count int) (*PortStateGetInfo, error) {
	bulkInfo, err := asicdClientMgr.ClientHdl.GetBulkPortState(asicdServices.Int(curMark), asicdServices.Int(count))
	if bulkInfo == nil {
		return nil, err
	}
	var portStateInfo PortStateGetInfo
	portStateInfo.StartIdx = int32(bulkInfo.StartIdx)
	portStateInfo.EndIdx = int32(bulkInfo.EndIdx)
	portStateInfo.Count = int32(bulkInfo.Count)
	portStateInfo.More = bulkInfo.More
	portStateInfo.PortStateList = make([]PortState, int(portStateInfo.Count))
	for idx := 0; idx < int(portStateInfo.Count); idx++ {
		portStateInfo.PortStateList[idx].PortNum = bulkInfo.PortStateList[idx].PortNum
		portStateInfo.PortStateList[idx].IfIndex = bulkInfo.PortStateList[idx].IfIndex
		portStateInfo.PortStateList[idx].Name = bulkInfo.PortStateList[idx].Name
		portStateInfo.PortStateList[idx].OperState = bulkInfo.PortStateList[idx].OperState
		portStateInfo.PortStateList[idx].NumUpEvents = bulkInfo.PortStateList[idx].NumUpEvents
		portStateInfo.PortStateList[idx].LastUpEventTime = bulkInfo.PortStateList[idx].LastUpEventTime
		portStateInfo.PortStateList[idx].NumDownEvents = bulkInfo.PortStateList[idx].NumDownEvents
		portStateInfo.PortStateList[idx].LastDownEventTime = bulkInfo.PortStateList[idx].LastDownEventTime
		portStateInfo.PortStateList[idx].Pvid = bulkInfo.PortStateList[idx].Pvid
		portStateInfo.PortStateList[idx].IfInOctets = bulkInfo.PortStateList[idx].IfInOctets
		portStateInfo.PortStateList[idx].IfInUcastPkts = bulkInfo.PortStateList[idx].IfInUcastPkts
		portStateInfo.PortStateList[idx].IfInDiscards = bulkInfo.PortStateList[idx].IfInDiscards
		portStateInfo.PortStateList[idx].IfInErrors = bulkInfo.PortStateList[idx].IfInErrors
		portStateInfo.PortStateList[idx].IfInUnknownProtos = bulkInfo.PortStateList[idx].IfInUnknownProtos
		portStateInfo.PortStateList[idx].IfOutOctets = bulkInfo.PortStateList[idx].IfOutOctets
		portStateInfo.PortStateList[idx].IfOutUcastPkts = bulkInfo.PortStateList[idx].IfOutUcastPkts
		portStateInfo.PortStateList[idx].IfOutDiscards = bulkInfo.PortStateList[idx].IfOutDiscards
		portStateInfo.PortStateList[idx].IfOutErrors = bulkInfo.PortStateList[idx].IfOutErrors
		portStateInfo.PortStateList[idx].ErrDisableReason = bulkInfo.PortStateList[idx].ErrDisableReason
	}
	return &portStateInfo, nil
}

func (asicdClientMgr *FSAsicdClientMgr) GetBulkVlanState(curMark, count int) (*VlanStateGetInfo, error) {
	bulkInfo, err := asicdClientMgr.ClientHdl.GetBulkVlanState(asicdServices.Int(curMark), asicdServices.Int(count))
	if bulkInfo == nil {
		return nil, err
	}
	var vlanStateInfo VlanStateGetInfo
	vlanStateInfo.StartIdx = int32(bulkInfo.StartIdx)
	vlanStateInfo.EndIdx = int32(bulkInfo.EndIdx)
	vlanStateInfo.Count = int32(bulkInfo.Count)
	vlanStateInfo.More = bulkInfo.More
	vlanStateInfo.VlanStateList = make([]VlanState, int(vlanStateInfo.Count))
	for idx := 0; idx < int(vlanStateInfo.Count); idx++ {
		vlanStateInfo.VlanStateList[idx].VlanId = bulkInfo.VlanStateList[idx].VlanId
		vlanStateInfo.VlanStateList[idx].VlanName = bulkInfo.VlanStateList[idx].VlanName
		vlanStateInfo.VlanStateList[idx].OperState = bulkInfo.VlanStateList[idx].OperState
		vlanStateInfo.VlanStateList[idx].IfIndex = bulkInfo.VlanStateList[idx].IfIndex
	}

	return &vlanStateInfo, nil
}

func (asicdClientMgr *FSAsicdClientMgr) GetBulkVlan(curMark, count int) (*VlanGetInfo, error) {
	bulkInfo, err := asicdClientMgr.ClientHdl.GetBulkVlan(asicdInt.Int(curMark), asicdInt.Int(count))
	if bulkInfo == nil {
		return nil, err
	}
	var vlanInfo VlanGetInfo
	vlanInfo.StartIdx = int32(bulkInfo.StartIdx)
	vlanInfo.EndIdx = int32(bulkInfo.EndIdx)
	vlanInfo.Count = int32(bulkInfo.Count)
	vlanInfo.More = bulkInfo.More
	vlanInfo.VlanList = make([]Vlan, int(vlanInfo.Count))
	for idx := 0; idx < int(vlanInfo.Count); idx++ {
		vlanInfo.VlanList[idx].VlanId = bulkInfo.VlanList[idx].VlanId
		vlanInfo.VlanList[idx].IfIndexList = append(vlanInfo.VlanList[idx].IfIndexList, bulkInfo.VlanList[idx].IfIndexList...)
		vlanInfo.VlanList[idx].UntagIfIndexList = append(vlanInfo.VlanList[idx].UntagIfIndexList, bulkInfo.VlanList[idx].UntagIfIndexList...)
	}
	return &vlanInfo, nil
}

func getAsicdThriftClientHdl(paramsFile string, logger *logging.Writer) *asicdServices.ASICDServicesClient {
	var asicdClient AsicdClient
	logger.Debug(fmt.Sprintln("Inside connectToServers...paramsFile", paramsFile))
	var clientsList []ClientJson

	bytes, err := ioutil.ReadFile(paramsFile)
	if err != nil {
		logger.Err("Error in reading configuration file")
		return nil
	}

	err = json.Unmarshal(bytes, &clientsList)
	if err != nil {
		logger.Err("Error in Unmarshalling Json")
		return nil
	}

	for _, client := range clientsList {
		if client.Name == "asicd" {
			logger.Debug(fmt.Sprintln("found asicd at port", client.Port))
			asicdClient.Address = "localhost:" + strconv.Itoa(client.Port)
			asicdClient.Transport, asicdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(asicdClient.Address)
			if err != nil {
				logger.Err(fmt.Sprintln("Failed to connect to Asicd, retrying until connection is successful"))
				count := 0
				ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
				for _ = range ticker.C {
					asicdClient.Transport, asicdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(asicdClient.Address)
					if err == nil {
						ticker.Stop()
						break
					}
					count++
					if (count % 10) == 0 {
						logger.Err("Still can't connect to Asicd, retrying..")
					}
				}

			}
			logger.Info("Connected to Asicd")
			asicdClient.ClientHdl = asicdServices.NewASICDServicesClientFactory(asicdClient.Transport, asicdClient.PtrProtocolFactory)
			return asicdClient.ClientHdl
		}
	}
	return nil
}
