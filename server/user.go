package main

import (
	"errors"
	"fmt"
	"regexp"
)

type UserInfo struct {
	Email  string `json:"email"`
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

func (p *Plugin) getEmailAndEmailName(mattermostUserId string) (string, string, error) {
	user, appErr := p.API.GetUser(mattermostUserId)
	if appErr != nil {
		p.errorf("error getting mattermost user from mattermostUserId: %s", mattermostUserId)
		return "", "", errors.New("error getting mattermost user from mattermostUserId, please contact your system administrator")
	}
	emailName, err := p.getUserFromEmail(user.Email)
	if err != nil {
		return "", "", err
	}

	return user.Email, emailName, nil
}

func (p *Plugin) getRoomOrDefault(mattermostUserId string) (string, error) {
	roomId, err := p.getRoom(mattermostUserId)
	if err == ErrUserNotFound {
		_, emailName, err2 := p.getEmailAndEmailName(mattermostUserId)
		if err2 != nil {
			return "", err2
		}
		return emailName, nil
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
