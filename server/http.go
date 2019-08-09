// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"net/http"
	"net/url"
	"strconv"
)

const (
	routeAPImeetings = "/api/v1/meetings"
	StatusStarted    = "STARTED"
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
	//config := p.getConfiguration()

	userId := r.Header.Get("Mattermost-User-Id")
	if userId == "" {
		return http.StatusUnauthorized, errors.New("not authorized")
	}

	var req startMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, fmt.Errorf("err: %v", err)
	}

	if _, appErr := p.API.GetChannelMember(req.ChannelID, userId); appErr != nil {
		return http.StatusForbidden, errors.New("forbidden")
	}

	roomId, err := p.getRoomOrDefault(userId)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	siteURL := p.getConfiguration().SiteHost
	if siteURL == "" {
		return http.StatusInternalServerError, errors.New("Unable to setup a meeting; the Webex plugin has not been configured yet.\nPlease ask your system administrator to set the `Webex Site Hostname` in `System Console -> PLUGINS -> Webex`.")
	}

	webexJoinURL := (&url.URL{
		Scheme: "https",
		Host:   siteURL,
		Path:   "/join/" + roomId,
	}).String()

	webexStartURL := (&url.URL{
		Scheme: "https",
		Host:   siteURL,
		Path:   "/start/" + roomId,
	}).String()

	joinPost := &model.Post{
		UserId:    p.botUserID,
		ChannelId: req.ChannelID,
		Message:   fmt.Sprintf("Meeting started at %s.", webexJoinURL),
		Type:      "custom_webex",
		Props: map[string]interface{}{
			"meeting_link":     webexJoinURL,
			"meeting_status":   StatusStarted,
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
