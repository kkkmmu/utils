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

package logging

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	nanomsg "github.com/op/go-nanomsg"
	"infra/sysd/sysdCommonDefs"
	"log"
	"log/syslog"
	"models/objects"
	"os"
	"sysd"
	"time"
)

const (
	DB_CONNECT_TIME_INTERVAL   = 2
	DB_CONNECT_RETRY_LOG_COUNT = 100
)

func ConvertLevelStrToVal(str string) sysdCommonDefs.SRDebugLevel {
	var val sysdCommonDefs.SRDebugLevel
	switch str {
	case "off":
		val = sysdCommonDefs.OFF
	case "crit":
		val = sysdCommonDefs.CRIT
	case "err":
		val = sysdCommonDefs.ERR
	case "warn":
		val = sysdCommonDefs.WARN
	case "alert":
		val = sysdCommonDefs.ALERT
	case "emerg":
		val = sysdCommonDefs.EMERG
	case "notice":
		val = sysdCommonDefs.NOTICE
	case "info":
		val = sysdCommonDefs.INFO
	case "debug":
		val = sysdCommonDefs.DEBUG
	case "trace":
		val = sysdCommonDefs.TRACE
	}
	return val
}

type Writer struct {
	sysLogger       *syslog.Writer
	nullLogger      *log.Logger
	GlobalLogging   bool
	MyComponentName string
	MyLogLevel      sysdCommonDefs.SRDebugLevel
	initialized     bool
	subSocket       *nanomsg.SubSocket
	socketCh        chan []byte
}

func NewLogger(name string, tag string, listenToConfig bool) (*Writer, error) {
	var err error
	srLogger := new(Writer)
	srLogger.MyComponentName = name
	srLogger.initialized = false

	srLogger.sysLogger, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, tag)
	if err != nil {
		fmt.Println("Failed to initialize syslog - ", err)
		return srLogger, err
	}
	// if sysLogger can't be initialized then send all logs to /dev/null
	devNull, err := os.Open(os.DevNull)
	if err == nil {
		srLogger.nullLogger = log.New(devNull, tag, log.Ldate|log.Ltime|log.Lshortfile)
	}

	srLogger.GlobalLogging = true
	srLogger.MyLogLevel = sysdCommonDefs.INFO
	// Read logging level from DB
	srLogger.readLogLevelFromDb()
	srLogger.initialized = true
	fmt.Println("Logging level ", srLogger.MyLogLevel, " set for ", srLogger.MyComponentName)
	if listenToConfig {
		go srLogger.ListenForLoggingNotifications()
	}
	return srLogger, err
}

func (logger *Writer) readSystemLoggingFromDb(dbHdl redis.Conn) error {
	logger.Info("Reading SystemLogging")
	var dbObj objects.SystemLogging
	objList, err := dbObj.GetAllObjFromDb(dbHdl)
	if err != nil {
		logger.Err("DB query failed for SystemLogging config")
		return err
	}
	if objList != nil {
		obj := sysd.NewSystemLogging()
		dbObject := objList[0].(objects.SystemLogging)
		objects.ConvertsysdSystemLoggingObjToThrift(&dbObject, obj)
		if obj.Logging == "on" {
			logger.GlobalLogging = true
		}
	}
	return nil
}

func (logger *Writer) readComponentLoggingFromDb(dbHdl redis.Conn) error {
	logger.Info("Reading ComponentLogging")
	var dbObj objects.ComponentLogging
	objList, err := dbObj.GetAllObjFromDb(dbHdl)
	if err != nil {
		logger.Err("DB query failed for ComponentLogging config")
		return err
	}
	if objList != nil {
		for idx := 0; idx < len(objList); idx++ {
			obj := sysd.NewComponentLogging()
			dbObject := objList[idx].(objects.ComponentLogging)
			objects.ConvertsysdComponentLoggingObjToThrift(&dbObject, obj)
			if obj.Module == logger.MyComponentName {
				logger.MyLogLevel = ConvertLevelStrToVal(obj.Level)
				return nil
			}
		}
	}
	return nil
}

func (logger *Writer) readLogLevelFromDb() error {
	var dbHdl redis.Conn
	var err error
	retryCount := 0
	ticker := time.NewTicker(DB_CONNECT_TIME_INTERVAL * time.Second)
	for _ = range ticker.C {
		retryCount += 1
		dbHdl, err = redis.Dial("tcp", ":6379")
		if err != nil {
			if retryCount%DB_CONNECT_RETRY_LOG_COUNT == 0 {
				logger.Err(fmt.Sprintln("Failed to dial out to Redis server. Ret    rying connection. Num retries = ", retryCount))
			}
		} else {
			break
		}
	}

	if dbHdl != nil {
		logger.readSystemLoggingFromDb(dbHdl)
		logger.readComponentLoggingFromDb(dbHdl)
		dbHdl.Close()
	}
	return nil
}

func (logger *Writer) SetGlobal(Enable bool) error {
	logger.GlobalLogging = Enable
	logger.Debug(fmt.Sprintln("Changed global logging to: ", logger.GlobalLogging, " for ", logger.MyComponentName))
	return nil
}

func (logger *Writer) SetLevel(level sysdCommonDefs.SRDebugLevel) error {
	logger.MyLogLevel = level
	logger.Debug(fmt.Sprintln("Changed logging level to: ", logger.MyLogLevel, " for ", logger.MyComponentName))
	return nil
}

func (logger *Writer) Crit(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.CRIT {
			return logger.sysLogger.Crit(message)
		}
	} else if logger.nullLogger != nil {
		logger.nullLogger.Println(message)
	}
	return nil
}

func (logger *Writer) Err(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.ERR {
			return logger.sysLogger.Err(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Warning(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.WARN {
			return logger.sysLogger.Warning(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Alert(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.ALERT {
			return logger.sysLogger.Alert(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Emerg(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.EMERG {
			return logger.sysLogger.Emerg(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Notice(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.NOTICE {
			return logger.sysLogger.Notice(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Info(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.INFO {
			return logger.sysLogger.Info(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Println(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.INFO {
			return logger.sysLogger.Info(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Debug(message string) error {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.DEBUG {
			return logger.sysLogger.Debug(message)
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return nil
}

func (logger *Writer) Write(message string) (int, error) {
	if logger.initialized {
		if logger.GlobalLogging && logger.MyLogLevel >= sysdCommonDefs.TRACE {
			n, err := logger.sysLogger.Write([]byte(message))
			return n, err
		} else if logger.nullLogger != nil {
			logger.nullLogger.Println(message)
		}
	}
	return 0, nil
}

func (logger *Writer) Close() error {
	var err error
	if logger.initialized {
		err = logger.sysLogger.Close()
	}
	logger = nil
	return err
}

func (logger *Writer) SetupSubSocket() error {
	var err error
	var socket *nanomsg.SubSocket
	if socket, err = nanomsg.NewSubSocket(); err != nil {
		logger.Err(fmt.Sprintf("Failed to create subscribe socket %s, error:%s", sysdCommonDefs.PUB_SOCKET_ADDR, err))
		return err
	}

	if err = socket.Subscribe(""); err != nil {
		logger.Err(fmt.Sprintf("Failed to subscribe to \"\" on subscribe socket %s, error:%s", sysdCommonDefs.PUB_SOCKET_ADDR, err))
		return err
	}

	if _, err = socket.Connect(sysdCommonDefs.PUB_SOCKET_ADDR); err != nil {
		logger.Err(fmt.Sprintf("Failed to connect to publisher socket %s, error:%s", sysdCommonDefs.PUB_SOCKET_ADDR, err))
		return err
	}

	logger.Info(fmt.Sprintf("Connected to publisher socket %s", sysdCommonDefs.PUB_SOCKET_ADDR))
	if err = socket.SetRecvBuffer(1024 * 1024); err != nil {
		logger.Err(fmt.Sprintln("Failed to set the buffer size for subsriber socket %s, error:", sysdCommonDefs.PUB_SOCKET_ADDR, err))
		return err
	}
	logger.subSocket = socket
	logger.socketCh = make(chan []byte)
	return nil
}

func (logger *Writer) ProcessLoggingNotification(rxBuf []byte) error {
	var msg sysdCommonDefs.Notification
	err := json.Unmarshal(rxBuf, &msg)
	if err != nil {
		logger.Err(fmt.Sprintln("Unable to unmarshal logging notification: ", rxBuf))
		return err
	}
	if msg.Type == sysdCommonDefs.G_LOG {
		var gLog sysdCommonDefs.GlobalLogging
		err = json.Unmarshal(msg.Payload, &gLog)
		if err != nil {
			logger.Err(fmt.Sprintln("Unable to unmarshal global logging notification: ", msg.Payload))
			return err
		}
		logger.SetGlobal(gLog.Enable)
	}
	if msg.Type == sysdCommonDefs.C_LOG {
		var cLog sysdCommonDefs.ComponentLogging
		err = json.Unmarshal(msg.Payload, &cLog)
		if err != nil {
			logger.Err(fmt.Sprintln("Unable to unmarshal component logging notification: ", msg.Payload))
			return err
		}
		if cLog.Name == logger.MyComponentName {
			logger.SetLevel(cLog.Level)
		}
	}
	return nil
}

func (logger *Writer) ProcessLogNotifications() error {
	for {
		select {
		case rxBuf := <-logger.socketCh:
			if rxBuf != nil {
				logger.ProcessLoggingNotification(rxBuf)
			}
		}
	}
	return nil
}

func (logger *Writer) ListenForLoggingNotifications() error {
	err := logger.SetupSubSocket()
	if err != nil {
		logger.Err(fmt.Sprintln("Failed to subscribe to logging notifications"))
		return err
	}
	go logger.ProcessLogNotifications()
	for {
		rxBuf, err := logger.subSocket.Recv(0)
		if err != nil {
			logger.Err(fmt.Sprintln("Recv on logging subscriber socket failed with error:", err))
			continue
		}
		logger.socketCh <- rxBuf
	}
	logger.Info(fmt.Sprintln("Existing logging config lister"))
	return nil
}
