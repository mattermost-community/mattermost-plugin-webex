package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

const (
	keyTokenSecret      = "token_secret"
	prefixOneTimeSecret = "ots_"  // + unique key that will be deleted after the first verification
	otsTTL              = 15 * 60 // expiry in seconds
)

type temporaryState struct {
	mmUserId string
	expires  time.Time
}

type store struct {
	plugin *Plugin
}

func NewStore(p *Plugin) Store {
	return &store{plugin: p}
}

type Store interface {
	OTSStore
}

type OTSStore interface {
	StoreTemporaryState(mmUserId, state string) error
	LoadTemporaryState(mmUserId string) (string, error)
}

func (store store) StoreTemporaryState(mmUserId, state string) error {
	tempState := &temporaryState{
		mmUserId: state,
		expires:  time.Now().Add(otsTTL * time.Second),
	}
	data, err := json.Marshal(&tempState)
	if err != nil {
		return err
	}

	appErr := store.plugin.API.KVSetWithExpiry(hashkey(prefixOneTimeSecret, mmUserId), data, otsTTL)
	if appErr != nil {
		return errors.WithMessage(appErr, "failed to store oauth temporary credentials for "+mmUserId)
	}
	return nil
}

func (store store) LoadTemporaryState(mmUserId string) (string, error) {
	b, appErr := store.plugin.API.KVGet(hashkey(prefixOneTimeSecret, mmUserId))
	if appErr != nil {
		return "", errors.WithMessage(appErr, "failed to load temporary credentials for "+mmUserId)
	}
	var tempState temporaryState
	err := json.Unmarshal(b, &tempState)
	if err != nil {
		return "", err
	}
	appErr = store.plugin.API.KVDelete(hashkey(prefixOneTimeSecret, mmUserId))
	if appErr != nil {
		return "", errors.WithMessage(appErr, "failed to delete temporary credentials for "+mmUserId)
	}
	if tempState.expires.Before(time.Now()) {
		return "", errors.New("Expired token")

	}
	return tempState.mmUserId, nil
}

func hashkey(prefix, key string) string {
	h := md5.New()
	_, _ = h.Write([]byte(key))
	return fmt.Sprintf("%s%x", prefix, h.Sum(nil))
}
