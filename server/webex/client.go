// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package webex

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type ClientError struct {
	StatusCode int
	Err        error
}

func (ce *ClientError) Error() string {
	return ce.Err.Error()
}

// Client represents a Webex API client
type Client struct {
	httpClient *http.Client
	xmlURL     string
	siteName   string
}

// NewClient returns a new Webex XML API client.
func NewClient(siteHost, siteName string) *Client {
	webexURL := (&url.URL{
		Scheme: "https",
		Host:   siteHost,
		Path:   "/WBXService/XMLService",
	}).String()

	return &Client{
		httpClient: &http.Client{},
		xmlURL:     webexURL,
		siteName:   siteName,
	}
}

// GetPersonalMeetingRoomUrl prefers roomId, username, and email for finding the PMR url (in that order).
func (c *Client) GetPersonalMeetingRoomUrl(roomId, username, email string) (string, *ClientError) {
	if roomId != "" {
		pmrUrl, err := c.getPMRFromRoomId(roomId)
		if err == nil && pmrUrl != "" {
			return pmrUrl, nil
		}
	}
	if username != "" {
		pmrUrl, err := c.getPMRFromUserName(username)
		if err == nil && pmrUrl != "" {
			return pmrUrl, nil
		}
	}
	if email != "" {
		pmrUrl, err := c.getPMRFromEmail(email)
		if err == nil && pmrUrl != "" {
			return pmrUrl, nil
		}
	}

	return "", &ClientError{500, errors.New("couldn't get PMR url")}
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

const roomIdContent = `<personalUrl>%s</personalUrl>`
const webexIdContent = `<webExId>%s</webExId>`
const emailContent = `<email>%s</email>`

func (c *Client) getPMRFromRoomId(roomId string) (string, *ClientError) {
	content := fmt.Sprintf(roomIdContent, roomId)
	payload := fmt.Sprintf(payloadWrapper, c.siteName, content)

	buf, cerr := c.request(payload)
	if cerr != nil {
		return "", cerr
	}

	var message GetPMRR
	err := xml.Unmarshal(buf.Bytes(), &message)
	if err != nil {
		return "", &ClientError{http.StatusInternalServerError, err}
	}

	return message.Body.BodyContent.PersonalMeetingRoom.PMRUrl, nil
}

func (c *Client) getPMRFromUserName(roomId string) (string, *ClientError) {
	return "", nil
}

func (c *Client) getPMRFromEmail(roomId string) (string, *ClientError) {
	return "", nil
}

func (c *Client) request(payload string) (*bytes.Buffer, *ClientError) {
	rq, err := http.NewRequest("POST", c.xmlURL, bytes.NewReader([]byte(payload)))
	if err != nil {
		return nil, &ClientError{http.StatusInternalServerError, err}
	}
	rq.Header.Set("Content-Type", "text/xml")
	rq.Close = true

	rp, err := c.httpClient.Do(rq)

	if err != nil {
		return nil, &ClientError{
			http.StatusInternalServerError,
			errors.WithMessagef(err, "failed request to %v", c.xmlURL),
		}
	}

	if rp == nil {
		return nil, &ClientError{
			http.StatusInternalServerError,
			errors.Errorf("received nil response when making request to %v", c.xmlURL),
		}
	}

	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return nil, &ClientError{
			rp.StatusCode,
			errors.New("Received status code above 300")}
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(rp.Body)
	if err != nil {
		return nil, &ClientError{
			http.StatusInternalServerError,
			errors.Errorf("Failed to read response from %v", c.xmlURL),
		}
	}

	return buf, nil
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		ioutil.ReadAll(r.Body)
		r.Body.Close()
	}
}
