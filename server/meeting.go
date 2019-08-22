package main

import (
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"github.com/mattermost/mattermost-server/model"
	"net/http"
	"strings"
)

// startMeeting can be used by the `/webex start` slash command or the http handleStartMeeting
// returns the joinPost, startPost, http status code and a descriptive error
func (p *Plugin) startMeeting(startedByUserId, meetingRoomOfUserId, channelId, meetingStatus string) (*model.Post, *model.Post, int, error) {
	roomUrl, err := p.getRoomUrlFromMMId(meetingRoomOfUserId)
	if err != nil {
		return nil, nil, http.StatusBadRequest, err
	}

	return p.startMeetingFromRoomUrl(roomUrl, startedByUserId, channelId, meetingStatus)
}

func (p *Plugin) startMeetingForUserId(header *model.CommandArgs, startedByUserId, meetingRoomOfUserId, meetingStatus string) error {
	if _, _, _, err := p.startMeeting(startedByUserId, meetingRoomOfUserId, header.ChannelId, meetingStatus); err != nil {
		return err
	}
	return nil
}

func (p *Plugin) startMeetingFromRoomUrl(roomUrl, startedByUserId, channelId, meetingStatus string) (*model.Post, *model.Post, int, error) {
	webexJoinURL := p.makeJoinUrl(roomUrl)
	webexStartURL := p.makeStartUrl(roomUrl)

	joinPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelId,
		Message:   fmt.Sprintf("Meeting started at %s.", webexJoinURL),
		Type:      "custom_webex",
		Props: map[string]interface{}{
			"meeting_link":     webexJoinURL,
			"meeting_status":   meetingStatus,
			"meeting_topic":    "Webex Meeting",
			"starting_user_id": startedByUserId,
		},
	}

	createdJoinPost, appErr := p.API.CreatePost(joinPost)
	if appErr != nil {
		return nil, nil, appErr.StatusCode, appErr
	}

	startPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelId,
		Message:   fmt.Sprintf("To start the meeting, click here: %s.", webexStartURL),
	}

	var createdStartPost *model.Post
	if meetingStatus == webex.StatusStarted {
		createdStartPost = p.API.SendEphemeralPost(startedByUserId, startPost)
	}

	return createdJoinPost, createdStartPost, http.StatusOK, nil
}

func (p *Plugin) makeJoinUrl(meetingUrl string) string {
	return strings.Replace(meetingUrl, "webex.com/meet/", "webex.com/join/", 1)
}

func (p *Plugin) makeStartUrl(meetingUrl string) string {
	return strings.Replace(meetingUrl, "webex.com/meet/", "webex.com/start/", 1)
}

func (p *Plugin) getUrlFromRoomId(roomId string) (string, error) {
	roomUrl, cerr := p.webexClient.GetPersonalMeetingRoomUrl(roomId, "", "")
	if cerr != nil {
		return "", cerr
	}

	return roomUrl, nil
}

func (p *Plugin) getUrlFromNameOrEmail(emailName, email string) (string, error) {
	roomUrl, cerr := p.webexClient.GetPersonalMeetingRoomUrl("", emailName, email)
	if cerr != nil {
		return "", cerr.Err
	}

	return roomUrl, nil
}

// getRoomUrlFromMMId will find the correct url for mattermostUserId, or return a message explaining why it couldn't.
func (p *Plugin) getRoomUrlFromMMId(mattermostUserId string) (string, error) {
	var roomUrl string
	if roomId, err := p.getRoom(mattermostUserId); err == nil && roomId != "" {
		// Look for their url using roomId
		roomUrl, err = p.getUrlFromRoomId(roomId)
		if err != nil {
			return "", fmt.Errorf("No Personal Room link found at `%s` for the room: `%s`", p.getConfiguration().SiteHost, roomId)
		}
	} else {
		// Look for their url using userName or email
		email, emailName, err := p.getEmailAndEmailName(mattermostUserId)
		if err != nil {
			return "", fmt.Errorf("Error getting email and emailName: %v", err)
		}
		roomUrl, err = p.getUrlFromNameOrEmail(emailName, email)
		if err != nil {
			return "", fmt.Errorf("No Personal Room link found at `%s` for your userName: `%s`, or your email: `%s`", p.getConfiguration().SiteHost, emailName, email)
		}
	}

	return roomUrl, nil
}
