package dbutils

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"models"
	"time"
	"utils/logging"
)

const (
	DB_CONNECT_TIME_INTERVAL   = 2
	DB_CONNECT_RETRY_LOG_COUNT = 100
)

type DBNotConnectedError struct {
	network string
	address string
}

func (e DBNotConnectedError) Error() string {
	return fmt.Sprintf("Not connected to DB at %s%s", e.network, e.address)
}

type DBUtil struct {
	redis.Conn
	logger  *logging.Writer
	network string
	address string
}

func NewDBUtil(logger *logging.Writer) *DBUtil {
	return &DBUtil{
		logger:  logger,
		network: "tcp",
		address: ":6379",
	}
}

func (db *DBUtil) Connect() error {
	retryCount := 0
	ticker := time.NewTicker(DB_CONNECT_TIME_INTERVAL * time.Second)
	for _ = range ticker.C {
		retryCount += 1
		dbHdl, err := redis.Dial(db.network, db.address)
		if err != nil {
			if retryCount%DB_CONNECT_RETRY_LOG_COUNT == 0 {
				if db.logger != nil {
					db.logger.Err(fmt.Sprintln("Failed to dial out to Redis server. Retrying connection. Num retries = ", retryCount))
				}
			}
		} else {
			db.Conn = dbHdl
			break
		}
	}
	return nil
}

func (db *DBUtil) Disconnect() {
	if db.Conn != nil {
		db.Close()
	}
}

func (db *DBUtil) StoreObjectInDb(obj models.ConfigObj) error {
	return obj.StoreObjectInDb(db.Conn)
}

func (db *DBUtil) DeleteObjectFromDb(obj models.ConfigObj) error {
	if db.Conn == nil {
		return DBNotConnectedError{db.network, db.address}
	}
	return obj.DeleteObjectFromDb(db.Conn)
}

func (db *DBUtil) GetObjectFromDb(obj models.ConfigObj, objKey string) (models.ConfigObj, error) {
	if db.Conn == nil {
		return obj, DBNotConnectedError{db.network, db.address}
	}
	return obj.GetObjectFromDb(objKey, db.Conn)
}

func (db *DBUtil) GetKey(obj models.ConfigObj) string {
	return obj.GetKey()
}

func (db *DBUtil) GetAllObjFromDb(obj models.ConfigObj) ([]models.ConfigObj, error) {
	if db.Conn == nil {
		return make([]models.ConfigObj, 0), DBNotConnectedError{db.network, db.address}
	}
	return obj.GetAllObjFromDb(db.Conn)
}

func (db *DBUtil) CompareObjectsAndDiff(obj models.ConfigObj, updateKeys map[string]bool, inObj models.ConfigObj) (
	[]bool, error) {
	if db.Conn == nil {
		return make([]bool, 0), DBNotConnectedError{db.network, db.address}
	}
	return obj.CompareObjectsAndDiff(updateKeys, inObj)
}

func (db *DBUtil) UpdateObjectInDb(obj, inObj models.ConfigObj, attrSet []bool) error {
	if db.Conn == nil {
		return DBNotConnectedError{db.network, db.address}
	}
	return obj.UpdateObjectInDb(inObj, attrSet, db.Conn)
}

func (db *DBUtil) MergeDbAndConfigObj(obj, dbObj models.ConfigObj, attrSet []bool) (models.ConfigObj, error) {
	return obj.MergeDbAndConfigObj(dbObj, attrSet)
}

func (db *DBUtil) GetBulkObjFromDb(obj models.ConfigObj, startIndex, count int64) (error, int64, int64, bool,
	[]models.ConfigObj) {
	if db.Conn == nil {
		return DBNotConnectedError{db.network, db.address}, 0, 0, false, make([]models.ConfigObj, 0)
	}
	return obj.GetBulkObjFromDb(startIndex, count, db.Conn)
}
