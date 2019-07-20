// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

const (
	routeAPImeetings      = "/api/v1/meetings"
	routeWebhook          = "/webhook"
	routeOauthRedirectURI = "/oauth"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

	switch path := r.URL.Path; path {
	case routeWebhook:
		p.handleWebhook(w, r)
	case routeAPImeetings:
		p.handleStartMeeting(w, r)
	case routeOauthRedirectURI:
		p.handleOauthRedirectURI(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleOauthRedirectURI(w http.ResponseWriter, r *http.Request) {
	p.Errorf("<><> request headers: %+v\n", r.Header)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		p.Errorf("error reading body")
	}
	r.ParseForm()
	p.Errorf("<><> form/query params: %+v", r.Form)
	p.Errorf("<><> request body: %s\n", b)
}

func (p *Plugin) handleWebhook(w http.ResponseWriter, r *http.Request) {
	//config := p.getConfiguration()
	//secret := config.WebhookSecret
	secret := ""
	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("secret")), []byte(secret)) != 1 {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request body", http.StatusBadRequest)
		return
	}

	var webhook webex.Webhook
	decoder := schema.NewDecoder()

	// Try to decode to standard webhook
	if err := decoder.Decode(&webhook, r.PostForm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.handleStandardWebhook(w, r, &webhook)

	// TODO: handle recording webhook
}

func (p *Plugin) handleStandardWebhook(w http.ResponseWriter, r *http.Request, webhook *webex.Webhook) {
	if webhook.Status != webex.WebhookStatusStarted {
		return
	}

	key := fmt.Sprintf("%v%v", postMeetingKey, webhook.ID)
	b, appErr := p.API.KVGet(key)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	if b == nil {
		return
	}
	postID := string(b)

	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	post.Message = "Meeting has ended."
	post.Props["meeting_status"] = webex.WebhookStatusEnded

	if _, appErr := p.API.UpdatePost(post); appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	if appErr := p.API.KVDelete(key); appErr != nil {
		p.API.LogWarn("failed to delete db entry", "error", appErr.Error())
		return
	}

	if _, err := w.Write([]byte(post.ToJson())); err != nil {
		p.API.LogWarn("failed to write response", "error", err.Error())
	}
}

type startMeetingRequest struct {
	ChannelID string `json:"channel_id"`
	Personal  bool   `json:"personal"`
	Topic     string `json:"topic"`
	MeetingID int    `json:"meeting_id"`
}

func (p *Plugin) handleStartMeeting(w http.ResponseWriter, r *http.Request) {
	//config := p.getConfiguration()

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req startMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	if _, appErr = p.API.GetChannelMember(req.ChannelID, userID); appErr != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	meetingID := req.MeetingID
	personal := req.Personal

	if meetingID == 0 && req.Personal {
		ru, clientErr := p.webexClient.GetUser(user.Email)
		if clientErr != nil {
			http.Error(w, clientErr.Error(), clientErr.StatusCode)
			return
		}
		meetingID = ru.Pmi
	}

	if meetingID == 0 {
		personal = false

		meeting := &webex.Meeting{
			Type:  webex.MeetingTypeInstant,
			Topic: req.Topic,
		}

		rm, clientErr := p.webexClient.CreateMeeting(meeting, user.Email)
		if clientErr != nil {
			http.Error(w, clientErr.Error(), clientErr.StatusCode)
			return
		}
		meetingID = rm.ID
	}

	//url := config.ZoomURL
	url := ""
	zoomURL := strings.TrimSpace(url)
	if len(zoomURL) == 0 {
		zoomURL = "https://zoom.us"
	}

	meetingURL := fmt.Sprintf("%s/j/%v", zoomURL, meetingID)

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: req.ChannelID,
		Message:   fmt.Sprintf("Meeting started at %s.", meetingURL),
		Type:      "custom_zoom",
		Props: map[string]interface{}{
			"meeting_id":       meetingID,
			"meeting_link":     meetingURL,
			"meeting_status":   webex.WebhookStatusStarted,
			"meeting_personal": personal,
			"meeting_topic":    req.Topic,
		},
	}

	createdPost, appErr := p.API.CreatePost(post)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	if appErr = p.API.KVSet(fmt.Sprintf("%v%v", postMeetingKey, meetingID), []byte(createdPost.Id)); appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("%v", meetingID))); err != nil {
		p.API.LogWarn("failed to write response", "error", err.Error())
	}
}
