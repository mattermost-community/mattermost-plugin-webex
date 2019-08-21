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

func (p *Plugin) getUrlFromRoomId(roomId string) (string, error) {
	roomUrl, cerr := p.webexClient.GetPersonalMeetingRoomUrl(roomId, "", "")
	if cerr != nil {
		return "", cerr
	}

	return roomUrl, nil
}

func (p *Plugin) getUrlFromEmail(mattermostUserId string) (string, error) {
	email, emailName, err := p.getEmailAndEmailName(mattermostUserId)
	if err != nil {
		return "", err
	}
	roomUrl, cerr := p.webexClient.GetPersonalMeetingRoomUrl("", emailName, email)
	if cerr != nil {
		return "", cerr.Err
	}

	return roomUrl, nil
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

// getRoomUrlFromMMId will find the correct url for mattermostUserId, or return a message explaining why it couldn't.
func (p *Plugin) getRoomUrlFromMMId(mattermostUserId string) (string, error) {
	email, emailName, err := p.getEmailAndEmailName(mattermostUserId)
	if err != nil {
		return "", fmt.Errorf("Error getting email and emailName: %v", err)
	}
	roomId, err := p.getRoom(mattermostUserId)
	if err == nil && roomId != "" {
		// Look for their url using roomId
		roomUrl, err := p.getUrlFromRoomId(roomId)
		if err != nil {
			return "", fmt.Errorf("No Personal Room link found at `%s` for the room: `%s`", p.getConfiguration().SiteHost, roomId)
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
