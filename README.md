# Mattermost Webex Plugin ![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-webex/master.svg)

Start and join voice calls, video calls and use screen sharing with your team members via Cisco Webex Cloud (This does not support Webex On-Prem).

Once enabled, clicking a video icon in a Mattermost channel invites team members to join a Webex meeting, the user who initiated the meeting will recive a "start" link in their chat window which prompts them to login and start the meeting for the other particpants.

## Requirements
- Mattermost Server 5.14+
- Paid Webex Cloud Account (Free Accounts do not work)

![image](https://user-images.githubusercontent.com)


## Configuration

1. Install the Mattermost Webex plugin by downloading the latest release from the [Releases Page](https://github.com/mattermost/mattermost-plugin-webex/releases) 
2. Go to **System Console > Plugins > Plugin Management** and select the tar.gz file and then click **Upload Plugin**
3. Go to **System Console > Plugins > Webex** and fill in the Webex Site Info with the webex domain name for your account.  Something like `https://mycompany.webex.com` 
4. Go to **System Console > Plugins > Webex** and click **Enable** button on the Webex plugin.  The plugin is now active for all users.

## Usage
1. The currently logged in user simply needs to press the "Webex" button on the top right corner in any mattermost Channel to start a webex meeting in their account.  The plugin will use the email address of the logged in Mattermost user to create the meeting link. Alternatively they can type in `/webex start`.
2. Any other users in the channel need to simply click the link and they will be brought to the "Lobby" of the meeting room and wait until the host has logged in and begun the meeting.

NOTE: For Users whose Webex login email is **different** than the email address they use to login to Mattermost:
- They will need to specify the name of their personal meeting room ID to the plugin
- For example if Bob's email address to login to Webex is bob@company.com, but in mattermost he logs in as bob.lastname@company.com, he will need to specify his Webex Personal Meeting Room ID manually before he can use the integration.  His personal meeting Room ID can be found in the webex account online and may be something like `bob`.  
- In which case he would set his personal meeting room by typing in `/webex room bob`


## Development

This plugin contains both a server and web app portion.

Use `make dist` to build distributions of the plugin that you can upload to a Mattermost server for testing.

Use `make check-style` to check the style for the whole plugin.

### Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within there, build the plugin like you would any other Go application.

### Web App

Inside the `/webapp` directory, you will find the JS and React files that make up the client-side of the plugin. Within there, modify files and components as necessary. Test your syntax by running `npm run build`.
