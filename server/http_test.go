// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlugin(t *testing.T) {
	noAuthMeetingRequest := httptest.NewRequest("POST", "/api/v1/meetings",
		strings.NewReader("{\"channel_id\": \"thechannelid\"}"))

	validMeetingRequest := httptest.NewRequest("POST", "/api/v1/meetings",
		strings.NewReader("{\"channel_id\": \"thechannelid\"}"))
	validMeetingRequest.Header.Add("Mattermost-User-Id", "theuserid")

	validMeetingRequest2 := httptest.NewRequest("POST", "/api/v1/meetings",
		strings.NewReader("{\"channel_id\": \"thechannelid\"}"))
	validMeetingRequest2.Header.Add("Mattermost-User-Id", "theuserid")

	validMeetingRequest3 := httptest.NewRequest("POST", "/api/v1/meetings",
		strings.NewReader("{\"channel_id\": \"thechannelid\"}"))
	validMeetingRequest3.Header.Add("Mattermost-User-Id", "theuserid")

	invalidMeetingRequestGet := httptest.NewRequest("GET", "/api/v1/meetings",
		strings.NewReader("{\"channel_id\": \"thechannelid\"}"))
	invalidMeetingRequestGet.Header.Add("Mattermost-User-Id", "theuserid")

	invalidMeetingRequestNoChannel := httptest.NewRequest("POST", "/api/v1/meetings",
		strings.NewReader("{\"channellll_id\": \"thechannelid\"}"))
	invalidMeetingRequestNoChannel.Header.Add("Mattermost-User-Id", "theuserid")

	for _, tc := range []struct {
		Name               string
		Request            *http.Request
		SiteHost           string
		ExpectedStatusCode int
	}{
		{
			Name:               "Unauthorized meeting request",
			Request:            noAuthMeetingRequest,
			SiteHost:           "hostname.webex.com",
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:               "Valid meeting request",
			Request:            validMeetingRequest,
			SiteHost:           "hostname.webex.com",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "No SiteHost set",
			Request:            validMeetingRequest2,
			SiteHost:           "",
			ExpectedStatusCode: http.StatusInternalServerError,
		},
		{
			Name:               "Invalid SiteHost set",
			Request:            validMeetingRequest3,
			SiteHost:           "blah.blah.webex.co",
			ExpectedStatusCode: http.StatusInternalServerError,
		},
		{
			Name:               "Invalid meeting request: using Get",
			Request:            invalidMeetingRequestGet,
			SiteHost:           "hostname.webex.com",
			ExpectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			Name:               "Invalid meeting request: no channel",
			Request:            invalidMeetingRequestNoChannel,
			SiteHost:           "hostname.webex.com",
			ExpectedStatusCode: http.StatusBadRequest,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			botUserID := "ason34aygl13nms0823nmastj3n99n"

			api := &plugintest.API{}

			api.On("GetChannelMember", "thechannelid", "theuserid").Return(&model.ChannelMember{}, nil)
			api.On("GetUser", "theuserid").Return(&model.User{Email: "theusername@thehost.com"}, nil)

			path, err := filepath.Abs("..")
			require.Nil(t, err)
			api.On("GetBundlePath").Return(path, nil)
			api.On("SetProfileImage", botUserID, mock.Anything).Return(nil)
			api.On("RegisterCommand", mock.Anything).Return(nil)
			api.On("LogDebug",
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string")).Return(nil)
			api.On("LogError",
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string"),
				mock.AnythingOfTypeArgument("string")).Return(nil)

			user, _ := json.Marshal(UserInfo{"myemail", "myroom"})
			api.On("KVGet", mock.AnythingOfTypeArgument("string")).Return(user, (*model.AppError)(nil))

			api.On("CreatePost",
				mock.AnythingOfType("*model.Post")).Return(&model.Post{}, nil)
			api.On("SendEphemeralPost",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*model.Post")).Return(nil)

			p := Plugin{}
			p.setConfiguration(&configuration{
				SiteHost: tc.SiteHost,
				siteName: parseSiteNameFromSiteHost(tc.SiteHost),
			})
			p.SetAPI(api)

			p.store = mockStore{}

			helpers := &plugintest.Helpers{}
			helpers.On("EnsureBot", mock.AnythingOfType("*model.Bot")).Return(botUserID, nil)
			p.SetHelpers(helpers)

			err = p.OnActivate()
			require.Nil(t, err)

			p.webexClient = webex.MockClient{SiteHost: tc.SiteHost}

			w := httptest.NewRecorder()

			p.ServeHTTP(&plugin.Context{}, w, tc.Request)
			assert.Equal(t, tc.ExpectedStatusCode, w.Result().StatusCode)

			if w.Result().StatusCode != http.StatusOK {
				return
			}

			webexJoinURL := "https://" + tc.SiteHost + "/join/myroom"
			expectedJoinPost := &model.Post{
				UserId:    p.botUserID,
				ChannelId: "thechannelid",
				Message:   fmt.Sprintf("Meeting started at %s.", webexJoinURL),
				Type:      "custom_webex",
				Props: map[string]interface{}{
					"meeting_link":     webexJoinURL,
					"meeting_status":   webex.StatusStarted,
					"meeting_topic":    "Webex Meeting",
					"starting_user_id": "theuserid",
				},
			}
			api.AssertCalled(t, "CreatePost", expectedJoinPost)

			webexStartURL := "https://" + tc.SiteHost + "/start/myroom"
			expectedStartPost := &model.Post{
				UserId:    p.botUserID,
				ChannelId: "thechannelid",
				Message:   fmt.Sprintf("To start the meeting, click here: %s.", webexStartURL),
			}
			api.AssertCalled(t, "SendEphemeralPost", "theuserid", expectedStartPost)
		})
	}
}
