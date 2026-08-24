package models

import (
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xo/dburl"
)

var log = &common.Log
var dbConn *dburl.URL
var supportedDB = []string{"mysql", "sqlite3"}

var (
	commonConnection *gorm.DB
	commonConnMu     sync.Mutex
)

func parseDBConnString() {
	var err error
	dbConn, err = dburl.Parse(common.DATABASE_URL)
	if err != nil {
		log.Fatal("Error parsing database connection ", common.DATABASE_URL, err)
	}
	_, ok := gorm.GetDialect(dbConn.Driver)
	if !ok || !funk.Contains(supportedDB, dbConn.Driver) {
		log.Fatal("Unsupported database: ", dbConn.Short())
	}
}

func GetDBConn() *dburl.URL {
	return dbConn
}

func SaveWithRetry(db *gorm.DB, i interface{}) error {
	var err error
	err = retry.Do(
		func() error {
			err = db.Save(i).Error
			if err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		log.Fatal("Failed to save ", err)
	}

	return nil
}

// GetDB returns the shared, pooled database handle. The connection is opened
// lazily exactly once and reused for the lifetime of the process; callers
// must NOT close it. (Previously every call opened a fresh connection and
// handlers closed it per request, which serialised player-polling workloads
// on connection setup.)
func GetDB() (*gorm.DB, error) {
	return GetCommonDB()
}

func GetCommonDB() (*gorm.DB, error) {
	if common.EnvConfig.DebugSQL {
		log.Debug("Getting Common DB handle from ", common.GetCallerFunctionName())
	}

	// Mutex-guarded lazy open that only caches on success — a transient
	// startup failure (e.g. MySQL not ready yet) is retried on the next call
	// instead of being cached forever.
	commonConnMu.Lock()
	defer commonConnMu.Unlock()

	if commonConnection != nil {
		return commonConnection, nil
	}

	err := retry.Do(
		func() error {
			conn, err := gorm.Open(dbConn.Driver, dbConn.DSN)
			if err != nil {
				return err
			}
			conn.LogMode(common.EnvConfig.DebugSQL)
			conn.DB().SetConnMaxIdleTime(4 * time.Minute)
			if common.DBConnectionPoolSize > 0 {
				conn.DB().SetMaxOpenConns(common.DBConnectionPoolSize)
			}
			commonConnection = conn
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return commonConnection, nil
}

// Lock functions

func CreateLock(lock string) {
	obj := KV{Key: "lock-" + lock, Value: "1"}
	obj.Save()

	common.PublishWS("lock.change", map[string]interface{}{"name": lock, "locked": true})
}

func CheckLock(lock string) bool {
	db, _ := GetDB()

	var obj KV
	err := db.Where(&KV{Key: "lock-" + lock}).First(&obj).Error

	return err == nil
}

func RemoveLock(lock string) {
	db, _ := GetDB()

	var obj KV
	db.Where(&KV{Key: "lock-" + lock}).Delete(&obj)

	common.PublishWS("lock.change", map[string]interface{}{"name": lock, "locked": false})
}

func RemoveAllLocks() {
	db, _ := GetDB()

	var locks []KV
	err := db.Where("`key` like 'lock-%'").Find(&locks).Error
	if err != nil {
		return
	}

	for _, lock := range locks {
		lockName := strings.Replace(lock.Key, "lock-", "", 1)
		RemoveLock(lockName)
	}
}

func init() {
	common.InitPaths()
	common.InitLogging()
	parseDBConnString()
	GetCommonDB()
}
