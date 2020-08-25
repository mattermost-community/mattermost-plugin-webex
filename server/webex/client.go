// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package webex

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const (
	StatusStarted = "STARTED"
	StatusInvited = "INVITED"
)

type Client interface {
	GetPersonalMeetingRoomURL(roomID, username, email string) (string, error)
}

// Client represents a Webex API client
type client struct {
	httpClient *http.Client
	xmlURL     string
	siteName   string
}

// NewClient returns a new Webex XML API client.
func NewClient(siteHost, siteName string) Client {
	webexURL := (&url.URL{
		Scheme: "https",
		Host:   siteHost,
		Path:   "/WBXService/XMLService",
	}).String()

	return &client{
		httpClient: &http.Client{},
		xmlURL:     webexURL,
		siteName:   siteName,
	}
}

// GetPersonalMeetingRoomURL prefers roomID, username, and email for finding the PMR url (in that order).
func (c *client) GetPersonalMeetingRoomURL(roomID, username, email string) (string, error) {
	if roomID != "" {
		pmrURL, err := c.getPMRFromRoomID(roomID)
		if err == nil && pmrURL != "" {
			return pmrURL, nil
		}
	}
	if username != "" {
		pmrURL, err := c.getPMRFromUserName(username)
		if err == nil && pmrURL != "" {
			return pmrURL, nil
		}
	}
	if email != "" {
		pmrURL, err := c.getPMRFromEmail(email)
		if err == nil && pmrURL != "" {
			return pmrURL, nil
		}
	}

	return "", errors.New("couldn't get PMR url")
}

const payloadWrapper = `<?xml version="1.0" encoding="UTF-8"?>
<serv:message xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <header>
        <securityContext>
            <siteName>%s</siteName>
        </securityContext>
    </header>
    <body>
        <bodyContent xsi:type="java:com.webex.service.binding.user.GetUserCard">
            %s
        </bodyContent>
    </body>
</serv:message>`

const roomIDContent = `<personalUrl>%s</personalUrl>`
const webexIDContent = `<webExId>%s</webExId>`
const emailContent = `<email>%s</email>`

// getPMRFromRoomID gets a Personal Meeting Room using a roomID, or returns an error if not found
func (c *client) getPMRFromRoomID(roomID string) (string, error) {
	content := fmt.Sprintf(roomIDContent, roomID)
	return c.getPMR(content)
}

// getPMRFromroomID gets a Personal Meeting Room using a userName, or returns an error if not found
func (c *client) getPMRFromUserName(userName string) (string, error) {
	content := fmt.Sprintf(webexIDContent, userName)
	return c.getPMR(content)
}

// getPMRFromroomID gets a Personal Meeting Room using an email, or returns an error if not found
func (c *client) getPMRFromEmail(email string) (string, error) {
	content := fmt.Sprintf(emailContent, email)
	return c.getPMR(content)
}

// getPMR gets a Personal Meeting Room given the body content
func (c *client) getPMR(content string) (string, error) {
	payload := fmt.Sprintf(payloadWrapper, c.siteName, content)
	buf, err := c.roundTrip(payload)
	if err != nil {
		return "", err
	}

	var message GetPMRR
	err = xml.Unmarshal(buf.Bytes(), &message)
	if err != nil {
		return "", err
	}

	return message.Body.BodyContent.PersonalMeetingRoom.PMRUrl, nil
}

func (c *client) roundTrip(payload string) (*bytes.Buffer, error) {
	rq, err := http.NewRequest("POST", c.xmlURL, bytes.NewReader([]byte(payload)))
	if err != nil {
		return nil, err
	}
	rq.Header.Set("Content-Type", "text/xml")
	rq.Close = true

	rp, err := c.httpClient.Do(rq)

	if err != nil {
		return nil, errors.WithMessagef(err, "failed request to %v", c.xmlURL)
	} else if rp == nil {
		return nil, errors.Errorf("received nil response when making request to %v", c.xmlURL)
	}

	defer rp.Body.Close()

	if rp.StatusCode >= 300 {
		return nil, errors.New("received status code above 300")
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(rp.Body)
	if err != nil {
		return nil, errors.Errorf("Failed to read response from %v", c.xmlURL)
	}

	return buf, nil
}

// For testing
type MockClient struct {
	SiteHost string
}

func (mc MockClient) GetPersonalMeetingRoomURL(roomID, username, email string) (string, error) {
	room := roomID
	if room == "" {
		room = username
	}
	if room == "" {
		room = getUserFromEmail(email)
	}
	return "https://" + mc.SiteHost + "/meet/" + room, nil
}

// only for testing
func getUserFromEmail(email string) string {
	ss := strings.Split(email, "@")
	if len(ss) != 2 {
		return ""
	}
	return ss[0]
}
