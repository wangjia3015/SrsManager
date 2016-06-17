package manager

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

const (
	//	TABLE_NAME_ROOM       = "room"
	TABLE_NAME_SRS_SERVER = "srs_server"
)

type DBSync struct {
	//db    *sql.DB
	dbDriver     string
	dbDataSource string
	dbName       string

	mutex sync.Mutex
}

func NewDBSync(dbDriver, dbDataSource, dbName string) *DBSync {
	return &DBSync{
		dbDriver:     dbDriver,
		dbDataSource: dbDataSource,
		dbName:       dbName,
	}
}

func (d *DBSync) open() (*sql.DB, error) {
	if db, err := sql.Open(d.dbDriver, d.dbDataSource); err != nil {
		return nil, err
	} else {
		return db, err
	}
}

func (d *DBSync) exec(sqlstr string, params ...interface{}) (sql.Result, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	var db *sql.DB
	var err error
	if db, err = d.open(); err != nil {
		return nil, err
	}
	defer db.Close()
	return db.Exec(sqlstr, params...)
}

func (d *DBSync) InsertRoom(room *Room) error {
	sql := "insert into room(`user`, `desc`, streamname, expiration, status, createtime, lastupdatetime) values(?, ?, ? , ?, ?, ?, ?)"

	room.CreateTime = time.Now().Unix()
	room.LastUpdateTime = room.CreateTime

	var id int64
	var err error
	if id, err = d.insert(sql,
		room.UserName,
		room.Desc,
		room.StreamName,
		room.Expiration,
		room.Status,
		room.CreateTime,
		room.LastUpdateTime,
	); err == nil {
		room.Id = id
	}
	return err
}

func (d *DBSync) insert(sql string, args ...interface{}) (lastInsertId int64, err error) {
	res, err := d.exec(sql, args...)
	if err != nil {
		return -1, fmt.Errorf("sql:%v args:%v insert err:%v", sql, args, err)
	}
	if lastInsertId, err = res.LastInsertId(); err != nil {
		return -1, fmt.Errorf("sql:%v args:%v getLastInsertId err:%v", sql, args, err)
	}

	return
}

func (d *DBSync) InsertEdge(e *Edge) error {
	sql := "insert into edge(`addr`,`role`,`desc`) values (?,?,?)"
	_, err := d.insert(sql, e.Addr, e.Role, e.Desc)

	return err
}

func (d *DBSync) UpdateRoom(room *Room) error {
	sql := "update room set `desc`= ?, `streamname`=? , `expiration` = ?, status = ?, `publishid` = ?,`publishhost` = ?, lastupdatetime=? where id = ?"
	room.LastUpdateTime = time.Now().Unix()
	if res, err := d.exec(sql,
		room.Desc,
		room.StreamName,
		room.Expiration,
		room.Status,
		room.PublishClientId,
		room.PublishHost,
		room.LastUpdateTime,
		room.Id); err == nil {
		if a, err := res.RowsAffected(); err != nil {
			glog.Warningln("res.RowsAffected", err)
		} else if a < 1 {
			glog.Warningln("res.RowsAffected row", a)
		}
		glog.Infoln("update room", room)
		return nil
	} else {
		return err
	}
}

func (d *DBSync) SelectRoom(params map[string]interface{}) (*Room, error) {
	keys := []string{}
	values := []interface{}{}

	for k, v := range params {
		keys = append(keys, " `"+k+"` = ? ")
		values = append(values, v)
	}

	sqlstr := "select `id`, `user`, `desc`, `streamname`, `expiration`, `status`, `publishid`, `publishhost`, `createtime`, `lastupdatetime` from room where " + strings.Join(keys, " and ")

	d.mutex.Lock()
	defer d.mutex.Unlock()
	var db *sql.DB
	var err error
	if db, err = d.open(); err != nil {
		return nil, err
	}
	defer db.Close()

	glog.Infoln(sqlstr, values[0])
	row := db.QueryRow(sqlstr, values...)

	var room Room
	if err = row.Scan(&room.Id,
		&room.UserName,
		&room.Desc,
		&room.StreamName,
		&room.Expiration,
		&room.Status,
		&room.PublishClientId,
		&room.PublishHost,
		&room.CreateTime,
		&room.LastUpdateTime); err != nil {
		return nil, err
	}

	return &room, nil
}

func (d *DBSync) InsertSrsServer(svr *SrsServer) error {
	sqlstr := "insert into " + TABLE_NAME_SRS_SERVER + "(`host`, `type`, `status`) values(?, ?, ?)"
	var err error
	svr.ID, err = d.insert(sqlstr, svr.Host, svr.ServerType, svr.Status)
	return err
}

func (d *DBSync) LoadSrsServers() ([]*SrsServer, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var db *sql.DB
	var err error
	if db, err = d.open(); err != nil {
		return nil, err
	}
	defer db.Close()

	sqlstr := "select `id`, `host`, `type`, `status` from " + TABLE_NAME_SRS_SERVER

	var rows *sql.Rows
	if rows, err = db.Query(sqlstr); err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*SrsServer
	for rows.Next() {
		var srs SrsServer
		if err = rows.Scan(&srs.ID,
			&srs.Host,
			&srs.ServerType,
			&srs.Status); err != nil {
			return nil, err
		}
		servers = append(servers, &srs)
	}
	return servers, nil
}
