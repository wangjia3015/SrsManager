package manager

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

type DBSync struct {
	//db    *sql.DB
	mutex        sync.Mutex
	dbDriver     string
	dbDataSource string
	dbName       string
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
	} else if err = d.useDB(db); err != nil {
		return nil, err
	} else {
		return db, err
	}
}

func (d DBSync) useDB(db *sql.DB) error {
	if _, err := db.Exec("use '" + d.dbName + "'"); err != nil {
		return err
	}
	return nil
}

func (d *DBSync) CreateTables() error {
	db, err := d.open()
	defer db.Close()
	if err != nil {
		return err
	}
	if err = d.useDB(db); err != nil {
		return err
	}

	sqls := []string{
		`CREATE TABLE 'room' (
			'id' bigint(20) NOT NULL AUTO_INCREMENT, 
			'user' varchar(255) NOT NULL, 
			'desc' varchar(255) NOT NULL, 
			'streamname' varchar(255) NOT NULL, 
			'expiration' int(11) NOT NULL, 
			'status' int NOT NULL, 
			'publishid'  int, 
			'publishhost' varchar(20) default NULL, 
			'lastupdatetime' int(11) NOT NULL, 
			'createtime' int(11) NOT NULL, 
			PRIMARY KEY ('id'), 
			UNIQUE KEY 'name' ('streamname') ) 
			ENGINE=InnoDB DEFAULT CHARSET=utf8 `,
	}
	for _, sql := range sqls {
		if _, err = db.Exec(sql); err != nil {
			glog.Warningln("sql err", err, sql)
			return err
		}
	}
	return nil
}

func (d *DBSync) Exec(sqlstr string, params ...interface{}) (sql.Result, error) {
	var db *sql.DB
	var err error
	if db, err = d.open(); err != nil {
		return nil, err
	}
	defer db.Close()
	stmt, err := db.Prepare(sqlstr)
	if err != nil {
		return nil, err
	}
	return stmt.Exec(params...)
}

func (d *DBSync) InsertRoom(room *Room) error {
	sql := "insert into room('user', 'desc', streamname, expiration, status, createtime, lastupdatetime) values(?, ?, ? , ?, ?, ?, ?)"

	room.CreateTime = time.Now().Unix()
	room.LastUpdateTime = room.CreateTime
	if res, err := d.Exec(sql, room.UserName,
		room.Desc,
		room.StreamName,
		room.Expiration,
		room.Status,
		room.CreateTime,
		room.LastUpdateTime,
	); err == nil {
		if room.Id, err = res.LastInsertId(); err != nil {
			glog.Warningln("res.LastInsertId err", err)
		}
		return err
	} else {
		return err
	}
}

func (d *DBSync) UpdateRoom(room *Room) error {
	sql := "update room set 'desc'= ?, 'streamname'=? , 'expiration' = ?, status = ?, 'publishid' = ?,'publishhost' = ?, lastupdatetime=? where id = ?"
	room.LastUpdateTime = time.Now().Unix()
	if res, err := d.Exec(sql,
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
		keys = append(keys, " '"+k+"' = ? ")
		values = append(values, v)
	}

	sqlstr := `select 'id', 'user', 'desc', 'streamname', 'expiration', 'status', 'publishid', 'publishhost', 'createtime', 'lastupdatetime' from room where ` + strings.Join(keys, " and ")

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
