package main

import (
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"github.com/mattermost/mattermost-server/v5/model"
	"net/http"
	"strings"
)

type meetingDetails struct {
	startedByUserId     string
	meetingRoomOfUserId string
	channelId           string
	meetingStatus       string
	roomUrl             string
}

type meetingPosts struct {
	createdJoinPost  *model.Post
	createdStartPost *model.Post
}

// startMeeting starts a meeting using details.meetingRoomOfUserId's room
// returns the joinPost, startPost, http status code and a descriptive error
func (p *Plugin) startMeeting(details meetingDetails) (*meetingPosts, int, error) {
	roomUrl, err := p.getRoomUrlFromMMId(details.meetingRoomOfUserId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	details.roomUrl = roomUrl
	return p.startMeetingFromRoomUrl(details)
}

// startMeetingFromRoomUrl starts a meeting using details.roomUrl, ignoring details.meetingRoomOfUserId
func (p *Plugin) startMeetingFromRoomUrl(details meetingDetails) (*meetingPosts, int, error) {
	webexJoinURL := p.makeJoinUrl(details.roomUrl)
	webexStartURL := p.makeStartUrl(details.roomUrl)

	joinPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: details.channelId,
		Message:   fmt.Sprintf("Meeting started at %s.", webexJoinURL),
		Type:      "custom_webex",
		Props: map[string]interface{}{
			"meeting_link":     webexJoinURL,
			"meeting_status":   details.meetingStatus,
			"meeting_topic":    "Webex Meeting",
			"starting_user_id": details.startedByUserId,
		},
	}

	createdJoinPost, appErr := p.API.CreatePost(joinPost)
	if appErr != nil {
		return nil, appErr.StatusCode, appErr
	}

	startPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: details.channelId,
		Message:   fmt.Sprintf("To start the meeting, click here: %s.", webexStartURL),
	}

	var createdStartPost *model.Post
	if details.meetingStatus == webex.StatusStarted {
		createdStartPost = p.API.SendEphemeralPost(details.startedByUserId, startPost)
	}

	return &meetingPosts{createdJoinPost, createdStartPost}, http.StatusOK, nil
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

func (p *Plugin) getUrlFromNameOrEmail(userName, email string) (string, error) {
	roomUrl, err := p.webexClient.GetPersonalMeetingRoomUrl("", userName, email)
	if err != nil {
		return "", err
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
		email, userName, err := p.getEmailAndUserName(mattermostUserId)
		if err != nil {
			return "", fmt.Errorf("Error getting email and Username: %v", err)
		}
		roomUrl, err = p.getUrlFromNameOrEmail(userName, email)
		if err != nil {
			return "", fmt.Errorf("No Personal Room link found at `%s` for your Username: `%s`, or your email: `%s`. Try setting a room manually with `/webex room <room id>`.", p.getConfiguration().SiteHost, userName, email)
		}
	}

	return roomUrl, nil
}
