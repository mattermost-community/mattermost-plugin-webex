package main

import (
	"fmt"
	"github.com/mattermost/mattermost-plugin-webex/server/webex"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"strings"
)

const helpText = "###### Mattermost Webex Plugin - Slash Command Help\n" +
	"* `/webex help` - This help text\n" +
	"* `/webex info` - Display your current settings\n" +
	"* `/webex start` - Start a Webex meeting in your room\n" +
	"* `/webex <room id>` - Shares a Join Meeting link for the Webex Personal Room meeting that is associated with the specified Personal Room ID, whether it’s your Personal Meeting Room ID or someone else’s.\n" +
	"* `/webex <@username>` - Shares a Join Meeting link for the Webex Personal Room meeting that is associated with that Mattermost team member.\n" +
	"###### Room Settings\n" +
	"* `/webex room <room id>` - Sets your personal Meeting Room ID. Meetings you start will use this ID. This setting is required only if your Webex account email address is different from your Mattermost account email address, or if the username of your email does not match your Personal Meeting Room ID or User name on your Webex site.\n" +
	"* `/webex room-reset` - Removes your room setting."

type CommandHandlerFunc func(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse

type CommandHandler struct {
	handlers       map[string]CommandHandlerFunc
	defaultHandler CommandHandlerFunc
}

var webexCommandHandler = CommandHandler{
	handlers: map[string]CommandHandlerFunc{
		"help":       executeHelp,
		"info":       executeInfo,
		"start":      executeStart,
		"room":       executeRoom,
		"room-reset": executeRoomReset,
	},
	defaultHandler: executeStartWithArg,
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

func executeHelp(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	return p.help(header)
}

func (p *Plugin) help(header *model.CommandArgs) *model.CommandResponse {
	p.postCommandResponse(header, helpText)
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
		AutoCompleteDesc: "Available commands: help, info, start, <room id/@username>, room, room-reset",
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
	if roomId == "" {
		roomId = "<not set>"
	}

	if len(args) != 1 {
		return p.responsef(header, "Please enter one new room id. Current room id is: `%s`", roomId)
	}

	userInfo, _ := p.store.LoadUserInfo(header.UserId)
	userInfo.RoomID = args[0]
	err = p.store.StoreUserInfo(header.UserId, userInfo)
	if err != nil {
		p.errorf("error in executeRoom: %v", err)
		return p.responsef(header, "Error storing user info, please contact your system administrator")
	}

	return p.responsef(header, "Room is set to: `%v`", userInfo.RoomID)
}

func executeRoomReset(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	userInfo, _ := p.store.LoadUserInfo(header.UserId)
	userInfo.RoomID = ""
	err := p.store.StoreUserInfo(header.UserId, userInfo)
	if err != nil {
		p.errorf("error in executeRoom: %v", err)
		return p.responsef(header, "Error storing user info, please contact your system administrator")
	}

	return p.responsef(header, "Room is set to: `<not set>`")
}

func executeInfo(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	roomId, err := p.getRoom(header.UserId)
	if err != nil && err != ErrUserNotFound {
		return p.responsef(header, err.Error())
	}
	if roomId == "" {
		roomId = "<not set>"
	}

	return p.responsef(header, "Webex site hostname: `%s`\nYour personal meeting room: `%s`", p.getConfiguration().SiteHost, roomId)
}

func executeStart(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if !p.getConfiguration().IsValid() {
		return p.responsef(header, "Unable to setup a meeting; the Webex plugin has not been configured correctly. Please contact your system administrator.")
	}

	details := meetingDetails{
		startedByUserId:     header.UserId,
		meetingRoomOfUserId: header.UserId,
		channelId:           header.ChannelId,
		meetingStatus:       webex.StatusStarted,
	}
	if _, _, err := p.startMeeting(details); err != nil {
		return p.responsef(header, err.Error())
	}
	return &model.CommandResponse{}
}

// executeStartWithArg looks for meeting urls given: room id, @username
func executeStartWithArg(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) != 1 {
		return p.help(header)
	}

	if !p.getConfiguration().IsValid() {
		return p.responsef(header, "Unable to setup a meeting; the Webex plugin has not been configured correctly. Please contact your system administrator.")
	}

	details := meetingDetails{
		startedByUserId: header.UserId,
		channelId:       header.ChannelId,
		meetingStatus:   webex.StatusInvited,
	}

	arg := args[0]
	if strings.HasPrefix(arg, "@") {
		// we were given a user
		user, appErr := p.API.GetUserByUsername(arg[1:])
		if appErr != nil {
			return p.responsef(header, "Could not find the user `%s`. Please make sure you typed the name correctly and try again.", arg)
		}
		details.meetingRoomOfUserId = user.Id
		if _, _, err := p.startMeeting(details); err != nil {
			return p.responsef(header, "Unable to create a meeting at `%s` for user: `%s`. They may not have their roomId set correctly, or their Mattermost email is not the same as their Webex email.", p.getConfiguration().SiteHost, arg)
		}
		return &model.CommandResponse{}
	}

	// we were given a roomId
	roomUrl, err := p.getUrlFromRoomId(arg)
	if err != nil {
		return p.responsef(header, "No Personal Room link found at `%s` for the room: `%s`", p.getConfiguration().SiteHost, arg)
	}

	details.roomUrl = roomUrl
	_, _, err = p.startMeetingFromRoomUrl(details)
	if err != nil {
		p.errorf("executeStartWithArg - Error creating the invitation posts, err: %v", err)
		return p.responsef(header, "Failed to make the invitation post. Please contact your system administrator.")
	}

	return &model.CommandResponse{}
}
