package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/command"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-webex/server/webex"
)

const helpText = "###### Mattermost Webex Plugin - Slash Command Help\n" +
	"* `/webex help` - This help text\n" +
	"* `/webex info` - Display your current settings\n" +
	"* `/webex start` - Start a Webex meeting in your room\n" +
	"* `/webex <room id>` - Shares a Join Meeting link for the Webex Personal Room meeting that is associated with the specified Personal Room ID, whether it’s your Personal Meeting Room ID or someone else’s.\n" +
	"* `/webex <@username>` - Shares a Join Meeting link for the Webex Personal Room meeting that is associated with that Mattermost team member.\n" +
	"###### Room Settings\n" +
	"* `/webex room <room id>` - Sets your personal Meeting Room ID. Meetings you start will use this ID. This setting is required only if your Webex account email address is different from your Mattermost account email address, or if the username of your email does not match your Personal Meeting Room ID or User name on your Webex site.\n" +
	"* `/webex room-reset` or `reset-room` - Removes your room setting."

const defaultRoomText = "not set (using your Mattermost email as the default)"

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
		"reset-room": executeRoomReset,
		"join":       executeStartWithArg, // Used as an alias for /webex <@username>/<room id> to allow for Autocomplete suggestions
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

	if !p.getConfiguration().IsValid() {
		return p.responsef(commandArgs, "The Webex plugin has not been configured correctly: the sitename has not been set. Please contact your system administrator."), nil
	}

	return webexCommandHandler.Handle(p, c, commandArgs, args[1:]...), nil
}

func (p *Plugin) getCommand() (*model.Command, error) {
	iconData, err := command.GetIconData(p.API, "assets/icon.svg")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get icon data")
	}

	return &model.Command{
		Trigger:              "webex",
		DisplayName:          "Webex",
		Description:          "Integration with Webex.",
		AutoComplete:         true,
		AutoCompleteDesc:     "Available commands: help, info, start, <room id/@username>, room, room-reset",
		AutoCompleteHint:     "[command]",
		AutocompleteData:     getAutocompleteData(),
		AutocompleteIconData: iconData,
	}, nil
}

func getAutocompleteData() *model.AutocompleteData {
	webexAutocomplete := model.NewAutocompleteData("webex", "[command]", "Available commands: help, info, start, <room id/@username>, room, room-reset")

	help := model.NewAutocompleteData("help", "", "Display usage information")
	webexAutocomplete.AddCommand(help)

	info := model.NewAutocompleteData("info", "", "Display your current settings")
	webexAutocomplete.AddCommand(info)

	start := model.NewAutocompleteData("start", "", "Start a Webex meeting in your room")
	webexAutocomplete.AddCommand(start)

	room := model.NewAutocompleteData("room", "<room id>", "Sets your personal Meeting Room ID")
	room.AddTextArgument("Webex meeting room ID", "<room id>", "")
	webexAutocomplete.AddCommand(room)

	roomReset := model.NewAutocompleteData("room-reset", "", "Removes your room setting")
	webexAutocomplete.AddCommand(roomReset)

	join := model.NewAutocompleteData("join", "<room id>/<@username>", "Shares a link to a Webex meeting in <room id> or in <@username>'s meeting room")
	join.AddTextArgument("Webex room ID or Mattermost username", "<room id>/<@username>", "")
	webexAutocomplete.AddCommand(join)

	return webexAutocomplete
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

func executeRoom(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	roomID, err := p.getRoomOrDefault(header.UserId)
	if err != nil {
		return p.responsef(header, err.Error())
	}
	if roomID == "" {
		roomID = defaultRoomText
	}

	if len(args) != 1 {
		return p.responsef(header, "Please enter one new room id. Current room id is: `%s`", roomID)
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

	return p.responsef(header, "Room is set to: `%s`", defaultRoomText)
}

func executeInfo(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	roomID, err := p.getRoom(header.UserId)
	if err != nil && err != ErrUserNotFound {
		fmt.Printf("<><> err: %+v type: %T", err, err)
		return p.responsef(header, err.Error())
	}
	if roomID == "" {
		roomID = defaultRoomText
	}

	return p.responsef(header, "Webex site hostname: `%s`\nYour personal meeting room: `%s`", p.getConfiguration().SiteHost, roomID)
}

func executeStart(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	details := meetingDetails{
		startedByUserID:     header.UserId,
		meetingRoomOfUserID: header.UserId,
		channelID:           header.ChannelId,
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

	details := meetingDetails{
		startedByUserID: header.UserId,
		channelID:       header.ChannelId,
		meetingStatus:   webex.StatusInvited,
	}

	arg := args[0]
	if strings.HasPrefix(arg, "@") {
		// we were given a user
		user, appErr := p.API.GetUserByUsername(arg[1:])
		if appErr != nil {
			return p.responsef(header, "Could not find the user `%s`. Please make sure you typed the name correctly and try again.", arg)
		}
		details.meetingRoomOfUserID = user.Id
		if _, _, err := p.startMeeting(details); err != nil {
			return p.responsef(header, "Unable to create a meeting at `%s` for user: `%s`. They may not have their roomID set correctly, or their Mattermost email is not the same as their Webex email.", p.getConfiguration().SiteHost, arg)
		}
		return &model.CommandResponse{}
	}

	// we were given a roomID
	roomURL, err := p.getURLFromRoomID(arg)
	if err != nil {
		return p.responsef(header, "No Personal Room link found at `%s` for the room: `%s`", p.getConfiguration().SiteHost, arg)
	}

	details.roomURL = roomURL
	_, _, err = p.startMeetingFromRoomURL(details)
	if err != nil {
		p.errorf("executeStartWithArg - Error creating the invitation posts, err: %v", err)
		return p.responsef(header, "Failed to make the invitation post. Please contact your system administrator.")
	}

	return &model.CommandResponse{}
}
