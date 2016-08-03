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
	"sort"
	"strconv"
	"strings"
	"time"
	"utils/commonDefs"
	"utils/logging"
	"utils/typeConv"
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
	StoreValInDb(interface{}, interface{}, interface{}) error
	GetAllKeys(interface{}) (interface{}, error)
	GetValFromDB(key interface{}, field interface{}) (interface{}, error)
}

type KeyObj struct {
	Key   string
	UTime int64
}

type KeyObjSlice []KeyObj

type RecvdEvent struct {
	eventId        events.EventId
	key            interface{}
	additionalInfo string
}

var GlobalEventEnable bool = true
var OwnerName string
var OwnerId events.OwnerId
var Logger logging.LoggerIntf
var PubHdl PubIntf
var PublishCh chan RecvdEvent

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

func InitEvents(ownerName string, pubHdl PubIntf, logger logging.LoggerIntf, evtChBufSize int32) error {
	EventMap = make(map[events.EventId]EventDetails)
	Logger = logger
	PubHdl = pubHdl
	Logger.Info(fmt.Sprintln("Initializing Owner Name :", ownerName))
	err := initEventDetails(ownerName)
	if err != nil {
		return err
	}
	PublishCh = make(chan RecvdEvent, evtChBufSize)
	go eventHandler()
	Logger.Info(fmt.Sprintln("EventMap:", EventMap))
	return nil
}

func eventHandler() {
	for {
		recvdEvt := <-PublishCh
		err := publishRecvdEvents(recvdEvt.eventId, recvdEvt.key, recvdEvt.additionalInfo)
		if err != nil {
			Logger.Err(fmt.Sprintln("Error Publishing Events:", err))
		}
	}
}

func PublishEvents(eventId events.EventId, key interface{}, additionalInfo string) error {
	recvdEvt := RecvdEvent{
		eventId:        eventId,
		key:            key,
		additionalInfo: additionalInfo,
	}
	PublishCh <- recvdEvt
	return nil
}

func publishRecvdEvents(eventId events.EventId, key interface{}, additionalInfo string) error {
	var err error
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
	if additionalInfo != "" {
		evt.Description = evtEnt.Description + ": " + additionalInfo
	} else {
		evt.Description = evtEnt.Description
	}
	evt.SrcObjName = evtEnt.SrcObjName
	evt.SrcObjKey = key
	msg, _ := json.Marshal(*evt)
	var unmarshalMsg events.Event
	err = json.Unmarshal(msg, &unmarshalMsg)
	keyMap, _ := events.EventKeyMap[evt.OwnerName]
	obj, _ := keyMap[evt.SrcObjName]
	obj = unmarshalMsg.SrcObjKey
	str := fmt.Sprintf("%v", obj)
	keyString := strings.TrimPrefix(str, "map[")
	strKey := strings.Split(keyString, "]")
	Logger.Info(fmt.Sprintln("Events to be published: ", evt, strKey[0]))
	keyStr := fmt.Sprintf("Events#%s#%s#%s#%s#%s#%d#", evt.OwnerName, evt.EventName, evt.SrcObjName, strKey[0], evt.TimeStamp.String(), evt.TimeStamp.UnixNano())
	Logger.Debug(fmt.Sprintln("Key Str :", keyStr))

	err = PubHdl.StoreValInDb(keyStr, evt.Description, "Desc")
	if err != nil {
		Logger.Err(fmt.Sprintln("Storing Events in database failed, err:", err))
	}
	PubHdl.Publish("PUBLISH", evt.OwnerName, msg)
	return nil
}

func GetEventQueryParams(r *http.Request) (evtObj events.EventObject, err error) {
	var body []byte
	if r != nil {
		body, err = ioutil.ReadAll(io.LimitReader(r.Body, commonDefs.MAX_JSON_LENGTH))
		if err != nil {
			return evtObj, err
		}
		if err = r.Body.Close(); err != nil {
			return evtObj, err
		}
	}

	if len(body) == 0 {
		return evtObj, err
	}
	err = json.Unmarshal(body, &evtObj)
	if err != nil {
		fmt.Println("UnmarshalObject returned error", err, "for ojbect info", evtObj)
	}
	return evtObj, err
}

func GetEvents(evtQueryObj events.EventObject, pubHdl PubIntf, logger logging.LoggerIntf) (evt []events.EventObject, err error) {
	qPattern := constructQueryPattern(evtQueryObj)
	//fmt.Println("Pattern Query:", qPattern)
	keys, err := typeConv.ConvertToStrings(pubHdl.GetAllKeys(qPattern))
	if err != nil {
		logger.Err(fmt.Sprintln("Error querying for keys:", err))
	}
	keySlice := constructKeySlice(keys)
	if keySlice == nil {
		logger.Err("Key slice is nil")
	}
	sort.Sort(KeyObjSlice(keySlice))
	for _, keyObj := range keySlice {
		desc, err := typeConv.ConvertToString(pubHdl.GetValFromDB(keyObj.Key, "Desc"))
		if err != nil {
			logger.Err(fmt.Sprintln("Error getting the value from DB", err))
			continue
		}
		str := strings.Split(keyObj.Key, "#")
		obj := events.EventObject{
			OwnerName:   str[1],
			EventName:   str[2],
			TimeStamp:   str[5],
			Description: desc,
			SrcObjName:  str[3],
			SrcObjKey:   str[4],
		}
		evt = append(evt, obj)
	}
	return evt, err
}

func constructQueryPattern(evtQueryObj events.EventObject) string {
	pattern := "Events#"
	if evtQueryObj.OwnerName == "" {
		pattern = pattern + "*#"
	} else {
		pattern = pattern + strings.ToUpper(evtQueryObj.OwnerName) + "#"
	}
	if evtQueryObj.EventName == "" {
		pattern = pattern + "*#"
	} else {
		pattern = pattern + evtQueryObj.EventName + "#"
	}
	if evtQueryObj.SrcObjName == "" {
		pattern = pattern + "*#"
	} else {
		pattern = pattern + evtQueryObj.SrcObjName + "#"
	}
	if evtQueryObj.SrcObjKey == "" {
		pattern = pattern + "*#"
	} else {
		pattern = pattern + evtQueryObj.SrcObjKey + "#"
	}
	pattern = pattern + "*"
	return pattern
}

func constructKeySlice(keys []string) []KeyObj {
	var kObjSlice []KeyObj
	for _, key := range keys {
		str := strings.Split(key, "#")
		uTime, err := strconv.ParseInt(str[len(str)-2], 10, 64)
		if err != nil {
			fmt.Println("Unable to Parse Int64")
			continue
		}
		kObj := KeyObj{
			Key:   key,
			UTime: uTime,
		}
		kObjSlice = append(kObjSlice, kObj)
	}
	return kObjSlice
}

func (kObjSlice KeyObjSlice) Less(i, j int) bool {
	return kObjSlice[i].UTime > kObjSlice[j].UTime
}

func (kObjSlice KeyObjSlice) Swap(i, j int) {
	kObjSlice[i], kObjSlice[j] = kObjSlice[j], kObjSlice[i]
}

func (kObjSlice KeyObjSlice) Len() int {
	return len(kObjSlice)
}
