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
	"io/ioutil"
	"models/events"
	"strings"
	"time"
	"utils/dbutils"
	"utils/logging"
)

type DaemonDetail struct {
	EventOwnerId   int
	EventOwnerName string
	EventEnable    bool
}

type EventJson struct {
	Daemons []DaemonDetail
}

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
var GlobalEventEnable bool = true
var OwnerName string
var OwnerId events.OwnerId
var Logger *logging.Writer
var DbHdl *dbutils.DBUtil

const (
	EventDir string = "/etc/flexswitch/"
)

func initOwnerDetails(ownerName string) error {
	var evtJson EventJson
	eventsFile := EventDir + "events.json"
	bytes, err := ioutil.ReadFile(eventsFile)
	if err != nil {
		Logger.Err(fmt.Sprintln("Error in reading ", eventsFile, " file."))
		err := errors.New(fmt.Sprintln("Error in rading ", eventsFile, " file."))
		return err
	}

	err = json.Unmarshal(bytes, &evtJson)
	if err != nil {
		Logger.Err(fmt.Sprintln("Errors in unmarshalling json file : ", eventsFile))
		err := errors.New(fmt.Sprintln("Errors in unmarshalling json file: ", eventsFile))
		return err
	}

	Logger.Info(fmt.Sprintln("Owner Name :", ownerName, "evtJson:", evtJson))
	for _, daemon := range evtJson.Daemons {
		Logger.Info(fmt.Sprintln("OwnerName:", ownerName, "damemon.OwnerName:", daemon.EventOwnerName))
		if daemon.EventOwnerName == ownerName {
			OwnerName = ownerName
			OwnerId = events.OwnerId(daemon.EventOwnerId)
			GlobalEventEnable = daemon.EventEnable
		}
	}

	Logger.Info(fmt.Sprintln("OwnerName:", OwnerName, "OwnerId:", OwnerId, "Enable:", GlobalEventEnable))
	return nil

}

func initEventList(ownerName string) error {
	var evts Events

	eventsFile := EventDir + strings.ToLower(ownerName) + "Events.json"
	bytes, err := ioutil.ReadFile(eventsFile)
	if err != nil {
		Logger.Err(fmt.Sprintln("Error in reading ", eventsFile, " file."))
		err := errors.New(fmt.Sprintln("Error in rading ", eventsFile, " file."))
		return err
	}

	err = json.Unmarshal(bytes, &evts)
	if err != nil {
		Logger.Err(fmt.Sprintln("Errors in unmarshalling json file : ", eventsFile))
		err := errors.New(fmt.Sprintln("Errors in unmarshalling json file: ", eventsFile))
		return err
	}

	for _, evt := range evts.EventList {
		evtEnt, exist := EventMap[events.EventId(evt.EventId)]
		if exist {
			Logger.Err(fmt.Sprintln("Duplicate event id :", evt.EventId))
			continue
		}

		evtEnt.Enable = evt.Enable
		evtEnt.Oid = OwnerId
		evtEnt.OwnerName = OwnerName
		evtEnt.Description = evt.Description
		evtEnt.EventName = evt.EventName
		evtEnt.SrcObjName = evt.SrcObjName
		EventMap[events.EventId(evt.EventId)] = evtEnt
	}

	Logger.Info(fmt.Sprintln("Event Map:", EventMap))
	return nil
}

func InitEvents(ownerName string, dbHdl *dbutils.DBUtil, logger *logging.Writer) error {

	EventMap = make(map[events.EventId]EventDetails)
	Logger = logger
	DbHdl = dbHdl
	Logger.Info(fmt.Sprintln("Initializing Owner Name :", ownerName))
	err := initOwnerDetails(ownerName)
	if err != nil {
		return err
	}

	err = initEventList(ownerName)
	if err != nil {
		return err
	}

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
	Logger.Info(fmt.Sprintln("Events to be published: ", evt))
	msg, _ := json.Marshal(*evt)
	DbHdl.Do("PUBLISH", evt.OwnerName, msg)
	return nil
}
