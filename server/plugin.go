// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"io/ioutil"
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
