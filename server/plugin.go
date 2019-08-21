// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/pkg/errors"
)

const (
	postMeetingKey = "post_meeting_"

	botUserName    = "webex"
	botDisplayName = "Webex"
	botDescription = "Created by the Webex plugin."
)

type Plugin struct {
	plugin.MattermostPlugin

	// botUserID of the created bot account.
	botUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// KV store
	store Store

	// the http client
	webexClient webex.Client
}

// OnActivate checks if the configurations is valid and ensures the bot account exists
func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()

	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    botUserName,
		DisplayName: botDisplayName,
		Description: botDescription,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	p.botUserID = botUserID

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	profileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", "profile.png"))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	if appErr := p.API.SetProfileImage(botUserID, profileImage); appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}

	p.store = NewStore(p)

	p.webexClient = webex.NewClient(config.SiteHost, config.siteName)

	err = p.API.RegisterCommand(getCommand())
	if err != nil {
		return errors.WithMessage(err, "OnActivate: failed to register command")
	}

	return nil
}

func (p *Plugin) GetPluginURLPath() string {
	return "/plugins/" + manifest.Id
}

func (p *Plugin) GetPluginURL() string {
	return strings.TrimRight(p.GetSiteURL(), "/") + p.GetPluginURLPath()
}

func (p *Plugin) GetSiteURL() string {
	return *p.API.GetConfig().ServiceSettings.SiteURL
}

func (p *Plugin) debugf(f string, args ...interface{}) {
	p.API.LogDebug(fmt.Sprintf(f, args...))
}

func (p *Plugin) infof(f string, args ...interface{}) {
	p.API.LogInfo(fmt.Sprintf(f, args...))
}

func (p *Plugin) errorf(f string, args ...interface{}) {
	p.API.LogError(fmt.Sprintf(f, args...))
}

func (p *Plugin) postEphemeralError(channelId, userId, msg string) {
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelId,
		Message:   msg,
	}

	_ = p.API.SendEphemeralPost(userId, post)
}

// startMeeting can be used by the `/webex start` slash command or the http handleStartMeeting
// returns the joinPost, startPost, http status code and a descriptive error
func (p *Plugin) startMeeting(startedByUserId, meetingRoomOfUserId, channelId, meetingStatus string) (*model.Post, *model.Post, int, error) {
	if !p.getConfiguration().IsValid() {
		return nil, nil, http.StatusInternalServerError, errors.New("Unable to setup a meeting; the Webex plugin has not been configured correctly.")
	}

	roomUrl, err := p.getRoomUrlFromMMId(meetingRoomOfUserId)
	if err != nil {
		return nil, nil, http.StatusBadRequest, err
	}

	return p.startMeetingFromRoomUrl(roomUrl, startedByUserId, channelId, meetingStatus)
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
