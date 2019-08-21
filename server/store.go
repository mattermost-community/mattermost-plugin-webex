package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

const (
	prefixUserInfo = "user_info_"
)

type Store interface {
	StoreUserInfo(mattermostUserId string, info UserInfo) error
	LoadUserInfo(mattermostUserId string) (UserInfo, error)
}

type store struct {
	plugin *Plugin
}

func NewStore(p *Plugin) Store {
	return &store{plugin: p}
}

func hashkey(prefix, key string) string {
	h := md5.New()
	_, _ = h.Write([]byte(key))
	return fmt.Sprintf("%s%x", prefix, h.Sum(nil))
}

func (store store) get(key string, v interface{}) (returnErr error) {
	defer func() {
		if returnErr == nil {
			return
		}
		returnErr = errors.WithMessage(returnErr, "failed to get from store")
	}()

	data, appErr := store.plugin.API.KVGet(key)
	if appErr != nil {
		return appErr
	}

	if data == nil {
		return nil
	}

	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

func (store store) set(key string, v interface{}) (returnErr error) {
	defer func() {
		if returnErr == nil {
			return
		}
		returnErr = errors.WithMessage(returnErr, "failed to store")
	}()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	appErr := store.plugin.API.KVSet(key, data)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (store store) StoreUserInfo(mattermostUserId string, info UserInfo) (returnErr error) {
	err := store.set(hashkey(prefixUserInfo, mattermostUserId), info)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("failed to store UserInfo for: %s", mattermostUserId))

	}
	return nil
}

var ErrUserNotFound = errors.New("user not found")

func (store store) LoadUserInfo(mattermostUserId string) (UserInfo, error) {
	userInfo := UserInfo{}
	err := store.get(hashkey(prefixUserInfo, mattermostUserId), &userInfo)
	if err != nil {
		return UserInfo{}, errors.WithMessage(err,
			fmt.Sprintf("failed to load userInfo for mattermostUserId: %s", mattermostUserId))
	}

	if len(userInfo.RoomID) == 0 {
		return UserInfo{}, ErrUserNotFound
	}

	return userInfo, nil

}
