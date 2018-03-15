package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

type Manager struct {
	dir      string
	filename string
}

func NewManager(dir, filename string) (*Manager, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	cfg := &Manager{dir: path.Join(usr.HomeDir, dir)}
	cfg.filename = path.Join(cfg.dir, filename)
	return cfg, nil
}

func (m *Manager) Prepare(i interface{}) error {
	if _, err := os.Stat(m.dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(m.dir, 0700); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if _, err := os.Stat(m.filename); err != nil {
		if os.IsNotExist(err) {
			return m.Save(i)
		}
		return err
	}
	return nil
}

func (m *Manager) Load(i interface{}) error {
	bytes, err := ioutil.ReadFile(m.filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, i); err != nil {
		return err
	}

	return nil
}

func (m *Manager) Save(i interface{}) error {
	bytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(m.filename, bytes, 0600)
}

func (m *Manager) Dir() string {
	return m.dir
}
