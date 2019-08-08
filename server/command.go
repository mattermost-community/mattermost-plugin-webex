package main

import (
	"crypto/rand"
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"net/http"
	"net/url"
	"strings"
)

const (
	webexOauthHost = "api.webex.com"
	webexOauthPath = "v1/oauth2/authorize"
)

const helpText = "###### Mattermost Webex Plugin - Slash Command Help\n" +
	"* `/webex connect` - Connect to your webex account\n"

type CommandHandlerFunc func(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse

type CommandHandler struct {
	handlers       map[string]CommandHandlerFunc
	defaultHandler CommandHandlerFunc
}

var jiraCommandHandler = CommandHandler{
	handlers: map[string]CommandHandlerFunc{
		"connect": executeConnect,
		"uri":     redirectURI,
	},
	defaultHandler: commandHelp,
}

func (ch CommandHandler) Handle(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	for n := len(args); n > 0; n-- {
		h := ch.handlers[strings.Join(args[:n], "/")]
		if h != nil {
			return h(p, c, header, args[n:]...)
		}
	}
	return ch.defaultHandler(p, c, header, args...)
}

func commandHelp(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	return p.help(header)
}

func (p *Plugin) help(args *model.CommandArgs) *model.CommandResponse {
	p.postCommandResponse(args, helpText)
	return &model.CommandResponse{}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	args := strings.Fields(commandArgs.Command)
	if len(args) == 0 || args[0] != "/webex" {
		return p.help(commandArgs), nil
	}
	return jiraCommandHandler.Handle(p, c, commandArgs, args[1:]...), nil
}

func executeConnect(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) != 0 {
		return p.help(header)
	}

	webexOauthURL := (&url.URL{
		Scheme: "https",
		Host:   webexOauthHost,
		Path:   "/" + webexOauthPath,
	}).String()

	req, err := http.NewRequest("GET", webexOauthURL, nil)
	if err != nil {
		p.errorf("Error building Oauth2 request, err: %v", err)
		return p.responsef(header, "Error connecting to Webex, please contact your system administrator.")
	}

	mattermostId := header.UserId
	randomBytes := make([]byte, 16)
	_, err = rand.Read(randomBytes)
	if err != nil {
		p.errorf("Error reading random bytes, err: %v", err)
		return p.responsef(header, "Error connecting to Webex, please contact your system administrator.")
	}

	state := fmt.Sprintf("%x", randomBytes)
	err = p.otsStore.StoreTemporaryState(mattermostId, state)
	if err != nil {
		p.errorf("Error storing temporary state, err: %v", err)
		return p.responsef(header, "Error connecting to Webex, please contact your system administrator.")
	}

	q := req.URL.Query()
	q.Add("response_type", "code")
	q.Add("client_id", p.getConfiguration().ClientID)
	q.Add("redirect_uri", p.GetPluginURL()+"/oauth")
	q.Add("scope", "all_read meeting_modify")
	q.Add("state", state)
	q.Add("code_challenge", "codechall")
	q.Add("code_challenge_method", "plain")

	req.URL.RawQuery = q.Encode()

	p.errorf("<><> redirect url: %s", req.URL.String())
	return p.responseRedirect(req.URL.String())
}

func redirectURI(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	return p.responsef(header, "redirectURI is: %s", p.GetPluginURL()+"/oauth")
}

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "webex",
		DisplayName:      "Webex",
		Description:      "Integration with Webex.",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect, help",
		AutoCompleteHint: "[command]",
	}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

func (p *Plugin) responsef(commandArgs *model.CommandArgs, format string, args ...interface{}) *model.CommandResponse {
	p.postCommandResponse(commandArgs, fmt.Sprintf(format, args...))
	return &model.CommandResponse{}
}

func (p *Plugin) responseRedirect(redirectURL string) *model.CommandResponse {
	return &model.CommandResponse{
		GotoLocation: redirectURL,
	}
}
