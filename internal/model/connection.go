package model

//Connection - is interface for getting access to database server
type Connection interface {
	IsExistedDatabase(dbName string) (bool)
	SizeDatabase(dbName string) (int64)
	GetURI() string 
}


