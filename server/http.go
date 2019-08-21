// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"net/http"
	"strconv"
)

const (
	routeAPImeetings = "/api/v1/meetings"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	status, err := handleHTTPRequest(p, w, r)
	if err != nil {
		p.API.LogError("ERROR: ", "Status", strconv.Itoa(status),
			"Error", err.Error(), "Host", r.Host, "RequestURI", r.RequestURI,
			"Method", r.Method, "query", r.URL.Query().Encode())
		http.Error(w, err.Error(), status)
		return
	}
	switch status {
	case http.StatusOK:
		// pass through
	case 0:
		status = http.StatusOK
	default:
		w.WriteHeader(status)
	}
	p.API.LogDebug("OK: ", "Status", strconv.Itoa(status), "Host", r.Host,
		"RequestURI", r.RequestURI, "Method", r.Method, "query", r.URL.Query().Encode())
}

func handleHTTPRequest(p *Plugin, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.URL.Path {
	case routeAPImeetings:
		return p.handleStartMeeting(w, r)
	}

	return http.StatusNotFound, errors.New("not found")
}

type startMeetingRequest struct {
	ChannelID string `json:"channel_id"`
	MeetingID int    `json:"meeting_id"`
}

func (p *Plugin) handleStartMeeting(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed,
			errors.New("method " + r.Method + " is not allowed, must be POST")
	}

	userId := r.Header.Get("Mattermost-User-Id")
	if userId == "" {
		return http.StatusUnauthorized, errors.New("not authorized")
	}

	if !p.getConfiguration().IsValid() {
		return http.StatusInternalServerError, errors.New("Unable to setup a meeting; the Webex plugin has not been configured correctly.")
	}

	var req startMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, fmt.Errorf("err: %v", err)
	}

	if req.ChannelID == "" {
		return http.StatusBadRequest, errors.New("channel id required")
	}

	if _, appErr := p.API.GetChannelMember(req.ChannelID, userId); appErr != nil {
		return http.StatusForbidden, errors.New("forbidden")
	}

	roomUrl, err := p.getRoomUrl(userId)
	if err != nil {
		p.postEphemeralError(req.ChannelID, userId, err.Error())
		return http.StatusBadRequest, nil
	}

	webexJoinURL := p.makeJoinUrl(roomUrl)
	webexStartURL := p.makeStartUrl(roomUrl)

	joinPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: req.ChannelID,
		Message:   fmt.Sprintf("Meeting started at %s.", webexJoinURL),
		Type:      "custom_webex",
		Props: map[string]interface{}{
			"meeting_link":     webexJoinURL,
			"meeting_status":   webex.StatusStarted,
			"meeting_topic":    "Webex Meeting",
			"starting_user_id": userId,
		},
	}

	createdPost, appErr := p.API.CreatePost(joinPost)
	if appErr != nil {
		return appErr.StatusCode, appErr
	}

	startPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: req.ChannelID,
		Message:   fmt.Sprintf("To start the meeting, click here: %s.", webexStartURL),
	}

	_ = p.API.SendEphemeralPost(userId, startPost)

	if _, err := w.Write([]byte(fmt.Sprintf("%v", createdPost.Id))); err != nil {
		p.API.LogWarn("failed to write response", "error", err.Error())
	}

	return http.StatusOK, nil
}
