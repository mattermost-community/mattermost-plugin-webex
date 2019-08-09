package main

import (
	"errors"
	"regexp"
)

type UserInfo struct {
	RoomID string `json:"room_id"`
}

func (p *Plugin) getUserFromEmail(email string) (string, error) {
	rexp := regexp.MustCompile("^(.*)@")
	matches := rexp.FindStringSubmatch(email)
	if matches == nil || matches[1] == "" {
		p.errorf("Error getting userName from email address, address: %v", email)
		return "", errors.New("error getting userName from email, please contact your system administrator")
	}

	return matches[1], nil
}

func (p *Plugin) getRoomOrDefault(mattermostUserId string) (string, error) {
	userInfo, err := p.store.LoadUserInfo(mattermostUserId)
	if err == ErrUserNotFound {
		// expected error
		user, appErr := p.API.GetUser(mattermostUserId)
		if appErr != nil {
			p.errorf("error getting mattermost user from mattermostUserId: %s", mattermostUserId)
			return "", errors.New("error getting mattermost user from mattermostUserId, please contact your system administrator")
		}

		// return the default
		roomId, err2 := p.getUserFromEmail(user.Email)
		if err2 != nil {
			return "", err2
		}
		return roomId, nil
	} else if err != nil {
		// unexpected error
		p.errorf("error from the store when retrieving room for mattermostUserId: %s, err: %v", mattermostUserId, err)
		return "", errors.New("error getting your room from the store, please contact your system administrator")
	}

	return userInfo.RoomID, nil
}
