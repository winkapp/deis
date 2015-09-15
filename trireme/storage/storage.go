/* Package storage defines a backend for storing and retrieving configuration data.

 */
package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var KeyNotFound = errors.New("Key not found")

type Storer interface {
	Get(ns, key string) (string, error)
	Set(ns, key, value string) error
	Remove(ns, key string) error
}

type storage map[string]map[string]string

type JSONStorage struct {
	loc string
	s   storage
}

func New(loc string) (*JSONStorage, error) {
	var s storage

	if _, err := os.Stat(loc); err != nil {
		if os.IsNotExist(err) {
			s := map[string]map[string]string{}
			return &JSONStorage{
				s:   s,
				loc: loc,
			}, nil
		} else {
			return nil, err
		}
	} else {
		data, err := ioutil.ReadFile(loc)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
	}

	return &JSONStorage{s: s, loc: loc}, nil
}

func (j *JSONStorage) Set(ns, key, value string) error {
	if _, ok := j.s[ns]; !ok {
		j.s[ns] = map[string]string{key: value}
		return j.Save()
	}
	j.s[ns][key] = value
	return j.Save()
}

func (j *JSONStorage) Remove(ns, key string) error {
	if inner, ok := j.s[ns]; ok {
		if _, ok := j.s[ns][key]; ok {
			delete(inner, key)
			if len(j.s[ns]) == 0 {
				delete(j.s, ns)
			}
			return j.Save()
		}
	}
	return KeyNotFound
}
func (j *JSONStorage) Get(ns, key string) (string, error) {
	if _, ok := j.s[ns]; ok {
		if val, ok := j.s[ns][key]; ok {
			return val, nil
		}
	}
	return "", KeyNotFound
}

func (j *JSONStorage) Save() error {
	data, err := json.Marshal(j.s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(j.loc, data, 0770)
}
