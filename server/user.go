package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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

func (p *Plugin) makeJoinUrl(meetingUrl string) string {
	return strings.Replace(meetingUrl, "webex.com/meet/", "webex.com/join/", 1)
}

func (p *Plugin) makeStartUrl(meetingUrl string) string {
	return strings.Replace(meetingUrl, "webex.com/meet/", "webex.com/start/", 1)
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

func (p *Plugin) getUrlFromRoom(mattermostUserId string) (string, error) {
	userInfo, err := p.store.LoadUserInfo(mattermostUserId)
	if err == ErrUserNotFound {
		return "", err
	}
	if err != nil {
		// unexpected error
		p.errorf("error from the store when retrieving room for mattermostUserId: %s, error: %v", mattermostUserId, err)
		return "", errors.New(fmt.Sprintf("error getting your room from the store, please contact your system administrator. Error: %v", err))
	}

	roomUrl, err := p.webexClient.GetPersonalMeetingRoomUrl(userInfo.RoomID, "", "")
	if err != nil {
		return "", err
	}

	return roomUrl, nil
}

func (p *Plugin) getUrlFromEmail(mattermostUserId string) (string, error) {
	email, emailName, err := p.getEmailAndEmailName(mattermostUserId)
	if err != nil {
		return "", err
	}
	roomUrl, err := p.webexClient.GetPersonalMeetingRoomUrl("", emailName, email)
	if err != nil {
		return "", err
	}

	return roomUrl, nil
}

func (p *Plugin) getRoomOrDefault(mattermostUserId string) (string, error) {
	userInfo, err := p.store.LoadUserInfo(mattermostUserId)
	if err == ErrUserNotFound {
		// expected error
		_, emailName, err2 := p.getEmailAndEmailName(mattermostUserId)
		if err2 != nil {
			return "", err2
		}
		return emailName, nil
	} else if err != nil {
		// unexpected error
		p.errorf("error from the store when retrieving room for mattermostUserId: %s, error: %v", mattermostUserId, err)
		return "", errors.New(fmt.Sprintf("error retrieving your room from the store, please contact your system administrator. Error: %v", err))
	}

	return userInfo.RoomID, nil
}

func (p *Plugin) getRoom(mattermostUserId string) (string, error) {
	userInfo, err := p.store.LoadUserInfo(mattermostUserId)
	if err != nil {
		return "", err
	}
	return userInfo.RoomID, nil
}

// getRoomUrl will find the correct url for mattermostUserId, or return a message explaining why it couldn't.
func (p *Plugin) getRoomUrl(mattermostUserId string) (string, error) {
	email, emailName, err := p.getEmailAndEmailName(mattermostUserId)
	if err != nil {
		return "", fmt.Errorf("Error getting email and emailName: %v", err)
	}
	roomId, err := p.getRoom(mattermostUserId)
	if err == nil && roomId != "" {
		// Look for their url using roomId
		roomUrl, err2 := p.getUrlFromRoom(mattermostUserId)
		if err2 != nil {
			return "", fmt.Errorf("No Personal Room link found at `%s` for your room: `%s`", p.getConfiguration().SiteHost, roomId)
		}

		return roomUrl, nil
	}

	// Look for their url using userName or email
	roomUrl, err := p.getUrlFromEmail(mattermostUserId)
	if err != nil {
		return "", fmt.Errorf("No Personal Room link found at `%s` for your userName: `%s`, or your email: `%s`", p.getConfiguration().SiteHost, emailName, email)
	}

	return roomUrl, nil

}
