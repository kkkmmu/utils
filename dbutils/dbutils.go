package dbutils

import (
	"fmt"
	"models"
	"utils/logging"

	"github.com/garyburd/redigo/redis"
)

type DBNotConnectedError struct {
	network string
	address string
}

func (e DBNotConnectedError) Error() string {
	return fmt.Sprintf("Not connected to DB at %s%s", e.network, e.address)
}

type DBUtil struct {
	dbHdl   redis.Conn
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
	dbHdl, err := redis.Dial(db.network, db.address)
	if err != nil {
		db.logger.Err("Failed to dial out to Redis server")
	} else {
		db.dbHdl = dbHdl
	}

	return err
}

func (db *DBUtil) Disconnect() {
	if db.dbHdl != nil {
		db.dbHdl.Close()
	}
}

func (db *DBUtil) StoreObjectInDb(obj models.ConfigObj) error {
	return obj.StoreObjectInDb(db.dbHdl)
}

func (db *DBUtil) DeleteObjectFromDb(obj models.ConfigObj) error {
	if db.dbHdl == nil {
		return DBNotConnectedError{db.network, db.address}
	}
	return obj.DeleteObjectFromDb(db.dbHdl)
}

func (db *DBUtil) GetObjectFromDb(obj models.ConfigObj, objKey string) (models.ConfigObj, error) {
	if db.dbHdl == nil {
		return obj, DBNotConnectedError{db.network, db.address}
	}
	return obj.GetObjectFromDb(objKey, db.dbHdl)
}

func (db *DBUtil) GetKey(obj models.ConfigObj) string {
	return obj.GetKey()
}

func (db *DBUtil) GetAllObjFromDb(obj models.ConfigObj) ([]models.ConfigObj, error) {
	if db.dbHdl == nil {
		return make([]models.ConfigObj, 0), DBNotConnectedError{db.network, db.address}
	}
	return obj.GetAllObjFromDb(db.dbHdl)
}

func (db *DBUtil) CompareObjectsAndDiff(obj models.ConfigObj, updateKeys map[string]bool, inObj models.ConfigObj) (
	[]bool, error) {
	if db.dbHdl == nil {
		return make([]bool, 0), DBNotConnectedError{db.network, db.address}
	}
	return obj.CompareObjectsAndDiff(updateKeys, inObj)
}

func (db *DBUtil) UpdateObjectInDb(obj, inObj models.ConfigObj, attrSet []bool) error {
	if db.dbHdl == nil {
		return DBNotConnectedError{db.network, db.address}
	}
	return obj.UpdateObjectInDb(inObj, attrSet, db.dbHdl)
}

func (db *DBUtil) MergeDbAndConfigObj(obj, dbObj models.ConfigObj, attrSet []bool) (models.ConfigObj, error) {
	return obj.MergeDbAndConfigObj(dbObj, attrSet)
}

func (db *DBUtil) GetBulkObjFromDb(obj models.ConfigObj, startIndex, count int64) (error, int64, int64, bool,
	[]models.ConfigObj) {
	if db.dbHdl == nil {
		return DBNotConnectedError{db.network, db.address}, 0, 0, false, make([]models.ConfigObj, 0)
	}
	return obj.GetBulkObjFromDb(startIndex, count, db.dbHdl)
}
