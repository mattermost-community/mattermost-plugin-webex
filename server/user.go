package main

import (
	"errors"
	"fmt"
)

type UserInfo struct {
	Email  string `json:"email"`
	RoomID string `json:"room_id"`
}

func (p *Plugin) getEmailAndUserName(mattermostUserID string) (string, string, error) {
	user, appErr := p.API.GetUser(mattermostUserID)
	if appErr != nil {
		p.errorf("error getting mattermost user from mattermostUserID: %s", mattermostUserID)
		return "", "", errors.New("error getting mattermost user from mattermostUserID, please contact your system administrator")
	}

	return user.Email, user.Username, nil
}

func (p *Plugin) getRoomOrDefault(mattermostUserID string) (string, error) {
	roomID, err := p.getRoom(mattermostUserID)
	if err == ErrUserNotFound {
		_, userName, err2 := p.getEmailAndUserName(mattermostUserID)
		if err2 != nil {
			return "", err2
		}
		return userName, nil
	} else if err != nil {
		return "", err
	}

	return roomID, nil
}

func (p *Plugin) getRoom(mattermostUserID string) (string, error) {
	userInfo, err := p.store.LoadUserInfo(mattermostUserID)
	if err == ErrUserNotFound {
		return "", err
	}
	if err != nil {
		// unexpected error
		p.errorf("error from the store when retrieving room for mattermostUserID: %s, error: %v", mattermostUserID, err)
		return "", fmt.Errorf("error getting your room from the store, please contact your system administrator. Error: %v", err)
	}
	return userInfo.RoomID, nil
}
