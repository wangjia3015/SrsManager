package manager

import (
	"sync"

	"github.com/orm"
)

type DBSync struct {
	mutex sync.Mutex
}

func NewDBSync(dbDriver, dbDataSource string) (*DBSync, error) {
	orm.RegisterDriver(dbDriver, orm.DR_MySQL)
	database := "default"
	orm.RegisterDataBase(database, dbDriver, dbDataSource)
	//orm.RegisterModel(new(NodesGroup))

	orm.SetMaxOpenConns(database, 10)
	orm.SetMaxIdleConns(database, 10)
	//create table
	forceCreate := false
	verbose := true
	err := orm.RunSyncdb(database, forceCreate, verbose)
	if err != nil {
		return nil, err
	}
	return &DBSync{}, nil
}
