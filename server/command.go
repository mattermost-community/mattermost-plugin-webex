package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"strings"
)

const helpText = "###### Mattermost Webex Plugin - Slash Command Help\n" +
	"* `/webex help` - This help text\n" +
	"* `/webex info` - Display your current settings\n" +
	"* `/webex room <room id>` - Set your room id. Meetings you start will use this id. Your room is currently set to: `%s`\n"

type CommandHandlerFunc func(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse

type CommandHandler struct {
	handlers       map[string]CommandHandlerFunc
	defaultHandler CommandHandlerFunc
}

var webexCommandHandler = CommandHandler{
	handlers: map[string]CommandHandlerFunc{
		"help": commandHelp,
		"info": executeInfo,
		"room": executeRoom,
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

func (p *Plugin) help(header *model.CommandArgs) *model.CommandResponse {
	roomId, err := p.getRoomOrDefault(header.UserId)
	if err != nil {
		return p.responsef(header, err.Error())
	}
	p.postCommandResponse(header, fmt.Sprintf(helpText, roomId))
	return &model.CommandResponse{}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	args := strings.Fields(commandArgs.Command)
	if len(args) == 0 || args[0] != "/webex" {
		return p.help(commandArgs), nil
	}
	return webexCommandHandler.Handle(p, c, commandArgs, args[1:]...), nil
}

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "webex",
		DisplayName:      "Webex",
		Description:      "Integration with Webex.",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: help, info, room",
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

func executeRoom(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	roomId, err := p.getRoomOrDefault(header.UserId)
	if err != nil {
		return p.responsef(header, err.Error())
	}
	if len(args) != 1 {
		return p.responsef(header, "Please enter one new room id. Current room id is: %s", roomId)
	}

	userInfo, _ := p.store.LoadUserInfo(header.UserId)
	userInfo.RoomID = args[0]
	err = p.store.StoreUserInfo(header.UserId, userInfo)
	if err != nil {
		p.errorf("error in executeRoom: %v", err)
		return p.responsef(header, "error storing user info, please contact your system administrator")
	}

	return p.responsef(header, "Room is set to: %v", userInfo.RoomID)
}

func executeInfo(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	roomId, err := p.getRoomOrDefault(header.UserId)
	if err != nil {
		return p.responsef(header, err.Error())
	}

	return p.responsef(header, "Webex site hostname: %s\nYour personal meeting room: %s", p.getConfiguration().SiteHost, roomId)
}
