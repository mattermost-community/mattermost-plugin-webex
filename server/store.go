package main

import (
	"crypto/md5" //nolint:gosec // md5 is used for user-hash generation and not encryption
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

const (
	prefixUserInfo = "user_info_"
)

type Store interface {
	StoreUserInfo(mattermostUserID string, info UserInfo) error
	LoadUserInfo(mattermostUserID string) (UserInfo, error)
}

type store struct {
	plugin *Plugin
}

func NewStore(p *Plugin) Store {
	return &store{plugin: p}
}

func hashkey(prefix, key string) string {
	h := md5.New() //nolint:gosec // md5 is used for user-hash generation and not encryption
	_, _ = h.Write([]byte(key))
	return fmt.Sprintf("%s%x", prefix, h.Sum(nil))
}

var ErrUserNotFound = errors.New("user not found")

func (store store) get(key string, v interface{}) error {
	data, appErr := store.plugin.API.KVGet(key)
	if appErr != nil {
		return appErr
	}

	if data == nil {
		return ErrUserNotFound
	}

	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

func (store store) set(key string, v interface{}) error {
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

func (store store) StoreUserInfo(mattermostUserID string, info UserInfo) error {
	// Set the email because we need a field in the userInfo that cannot be blank (in order to tell if a user was found)
	email, _, err := store.plugin.getEmailAndUserName(mattermostUserID)
	if err != nil {
		return err
	}
	info.Email = email
	err = store.set(hashkey(prefixUserInfo, mattermostUserID), info)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("failed to store UserInfo for: %s", mattermostUserID))
	}
	return nil
}

func (store store) LoadUserInfo(mattermostUserID string) (UserInfo, error) {
	userInfo := UserInfo{}
	err := store.get(hashkey(prefixUserInfo, mattermostUserID), &userInfo)
	if err != nil && err == ErrUserNotFound {
		return UserInfo{}, err
	}
	if err != nil {
		return UserInfo{}, errors.WithMessage(err,
			fmt.Sprintf("failed to load userInfo for mattermostUserId: %s", mattermostUserID))
	}

	if len(userInfo.Email) == 0 {
		return UserInfo{}, ErrUserNotFound
	}
	return userInfo, nil
}
