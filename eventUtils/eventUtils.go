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

package eventUtils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"models/events"
	"net/http"
	"time"
	"utils/logging"
)

const (
	MAX_JSON_LENGTH = 4096
)

type Event struct {
	EventId     int
	EventName   string
	Description string
	SrcObjName  string
	Enable      bool
}

type Events struct {
	EventList []Event
}

type EventDetails struct {
	Enable      bool
	Oid         events.OwnerId
	OwnerName   string
	EventName   string
	Description string
	SrcObjName  string
}

var EventMap map[events.EventId]EventDetails

type FaultDetail struct {
	RaiseFault       bool
	ClearingEventId  int
	ClearingDaemonId int
}

type EventStruct struct {
	EventId     int
	EventName   string
	Description string
	SrcObjName  string
	EventEnable bool
	IsFault     bool
	Fault       FaultDetail
}

type DaemonEvent struct {
	DaemonId          int
	DaemonName        string
	DaemonEventEnable bool
	EventList         []EventStruct
}

type EventJson struct {
	DaemonEvents []DaemonEvent
}

type PubIntf interface {
	Publish(string, interface{}, interface{})
	StoreEventsInDb(string, string) error
}

var GlobalEventEnable bool = true
var OwnerName string
var OwnerId events.OwnerId
var Logger logging.LoggerIntf
var PubHdl PubIntf

const (
	EventDir string = "/etc/flexswitch/"
)

func initEventDetails(ownerName string) error {
	var evtJson EventJson
	eventsFile := EventDir + "events.json"
	bytes, err := ioutil.ReadFile(eventsFile)
	if err != nil {
		Logger.Err(fmt.Sprintln("Error in reading ", eventsFile, " file."))
		err := errors.New(fmt.Sprintln("Error in reading ", eventsFile, " file."))
		return err
	}

	err = json.Unmarshal(bytes, &evtJson)
	if err != nil {
		Logger.Err(fmt.Sprintln("Errors in unmarshalling json file : ", eventsFile))
		err := errors.New(fmt.Sprintln("Errors in unmarshalling json file: ", eventsFile))
		return err
	}

	Logger.Debug(fmt.Sprintln("Owner Name :", ownerName, "evtJson:", evtJson))
	for _, daemon := range evtJson.DaemonEvents {
		Logger.Debug(fmt.Sprintln("OwnerName:", ownerName, "daemon.DaemonName:", daemon.DaemonName))
		if daemon.DaemonName == ownerName {
			OwnerName = ownerName
			OwnerId = events.OwnerId(daemon.DaemonId)
			GlobalEventEnable = daemon.DaemonEventEnable
			for _, evt := range daemon.EventList {
				evtId := events.EventId(evt.EventId)
				evtEnt, _ := EventMap[evtId]
				evtEnt.EventName = evt.EventName
				evtEnt.Description = evt.Description
				evtEnt.SrcObjName = evt.SrcObjName
				evtEnt.Oid = OwnerId
				evtEnt.OwnerName = OwnerName
				evtEnt.Enable = evt.EventEnable
				EventMap[evtId] = evtEnt
			}
			continue
		}
	}

	return nil
}

func InitEvents(ownerName string, pubHdl PubIntf, logger logging.LoggerIntf) error {

	EventMap = make(map[events.EventId]EventDetails)
	Logger = logger
	PubHdl = pubHdl
	Logger.Info(fmt.Sprintln("Initializing Owner Name :", ownerName))
	err := initEventDetails(ownerName)
	if err != nil {
		return err
	}

	Logger.Info(fmt.Sprintln("EventMap:", EventMap))
	return nil
}

func PublishEvents(eventId events.EventId, key interface{}) error {
	if GlobalEventEnable == false {
		return nil
	}
	evt := new(events.Event)
	evtEnt, exist := EventMap[eventId]
	if !exist {
		err := errors.New(fmt.Sprintln("Unable to find the event corresponding to given eventId: ", eventId))
		return err
	}

	if evtEnt.Enable == false {
		return nil
	}
	evt.OwnerId = evtEnt.Oid
	evt.OwnerName = evtEnt.OwnerName
	evt.EvtId = eventId
	evt.EventName = evtEnt.EventName
	evt.TimeStamp = time.Now()
	evt.Description = evtEnt.Description
	evt.SrcObjName = evtEnt.SrcObjName
	evt.SrcObjKey = key
	Logger.Debug(fmt.Sprintln("Events to be published: ", evt))
	keyStr := fmt.Sprintf("Events#%s#%s#%s#%s#", evt.OwnerName, evt.SrcObjName, evt.SrcObjKey, evt.TimeStamp.String())
	Logger.Debug(fmt.Sprintln("Key Str :", keyStr))
	err := PubHdl.StoreEventsInDb(keyStr, evt.Description)
	if err != nil {
		Logger.Err(fmt.Sprintln("Storing Events in database failed, err:", err))
	}
	msg, _ := json.Marshal(*evt)
	PubHdl.Publish("PUBLISH", evt.OwnerName, msg)
	return nil
}

func GetEventQueryParams(r *http.Request) (evtObj events.EventObject, err error) {
	var body []byte
	if r != nil {
		body, err = ioutil.ReadAll(io.LimitReader(r.Body, MAX_JSON_LENGTH))
		if err != nil {
			return evtObj, err
		}
		if err = r.Body.Close(); err != nil {
			return evtObj, err
		}
	}

	err = json.Unmarshal(body, &evtObj)
	if err != nil {
		fmt.Println("UnmarshalObject returnexd error", err, "for ojbect info", evtObj)
	}
	return evtObj, err
}

func GetEvents(evtQueryObj events.EventObject, pubHdl PubIntf, logger logging.LoggerIntf) (evt []events.EventObject, err error) {
	fmt.Println("Event Query Object:", evtQueryObj)
	obj := events.EventObject{
		OwnerName:   "ASICD",
		EventName:   "PortStateUp",
		TimeStamp:   time.Now().String(),
		Description: "Front Panel Port Went UP",
		SrcObjName:  "Port",
		SrcObjKey:   "{IntfRef:fpPort1}",
	}

	evt = append(evt, obj)
	return evt, err
}
