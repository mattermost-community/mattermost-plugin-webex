package main

import (
	"errors"
	"fmt"
)

type UserInfo struct {
	Email  string `json:"email"`
	RoomID string `json:"room_id"`
}

func (p *Plugin) getEmailAndUserName(mattermostUserId string) (string, string, error) {
	user, appErr := p.API.GetUser(mattermostUserId)
	if appErr != nil {
		p.errorf("error getting mattermost user from mattermostUserId: %s", mattermostUserId)
		return "", "", errors.New("error getting mattermost user from mattermostUserId, please contact your system administrator")
	}

	return user.Email, user.Username, nil
}

func (p *Plugin) getRoomOrDefault(mattermostUserId string) (string, error) {
	roomId, err := p.getRoom(mattermostUserId)
	if err == ErrUserNotFound {
		_, userName, err2 := p.getEmailAndUserName(mattermostUserId)
		if err2 != nil {
			return "", err2
		}
		return userName, nil
	} else if err != nil {
		return "", err
	}

	return roomId, nil
}

func (p *Plugin) getRoom(mattermostUserId string) (string, error) {
	userInfo, err := p.store.LoadUserInfo(mattermostUserId)
	if err == ErrUserNotFound {
		return "", err
	}
	if err != nil {
		// unexpected error
		p.errorf("error from the store when retrieving room for mattermostUserId: %s, error: %v", mattermostUserId, err)
		return "", errors.New(fmt.Sprintf("error getting your room from the store, please contact your system administrator. Error: %v", err))
	}
	return userInfo.RoomID, nil
}
