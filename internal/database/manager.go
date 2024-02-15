package database

import "github.com/pkg/errors"

type configManager interface {
	GetAsSliceOfMaps() map[string]map[string]any
}

type Manager struct {
	host        string
	port        int
	user        string
	password    string
	dbInterface string
}

type Managers map[string]Manager

func (e Managers) GetById(id string) Manager {
	return e[id]
}

func (m Manager) GetAuth() (name string, password string) {
	name, password = m.user, m.password
	return
}

func (m Manager) GetSocket() (host string, port int) {
	host, port = m.host, m.port
	return
}

func InitManagersFromConfig(config configManager) (Managers, error) {
	res := make(Managers)
	for k, v := range config.GetAsSliceOfMaps() {

		m := Manager{}

		if v, err := getValueAsString(v["host"]); err == nil {
			m.host = v
		} else {
			return nil, err
		}

		if v, err := getValueAsInt(v["port"]); err == nil {
			m.port = v
		} else {
			return nil, err
		}

		if v, err := getValueAsString(v["user"]); err == nil {
			m.user = v
		} else {
			return nil, err
		}

		if v, err := getValueAsString(v["password"]); err == nil {
			m.password = v
		} else {
			return nil, err
		}

		if v, err := getValueAsString(v["dbInterface"]); err == nil {
			m.dbInterface = v
		} else {
			return nil, err
		}

		res[k] = m
	}
	return res, nil
}

func getValueAsString(a any) (string, error) {
	switch v := a.(type) {
	case string:
		return v, nil
	default:
		return "", errors.Errorf("value is not string. %v = %T", v, v)
	}
}

func getValueAsInt(a any) (int, error) {
	switch v := a.(type) {
	case int:
		return v, nil
	default:
		return 0, errors.Errorf("value is not string. %v = %T", v, v)
	}
}
