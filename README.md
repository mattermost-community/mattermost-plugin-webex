# Mattermost Webex Cloud Plugin ![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-webex/master.svg)

Start and join voice calls, video calls and use screen sharing with your team members via Cisco Webex Meetings.

Once enabled, clicking a meeting icon in a Mattermost channel invites team members to join a Webex meeting, hosted using the credentials of the user who initiated the call.

![image](https://user-images.githubusercontent.com/915956/65968532-2bff2400-e418-11e9-8479-3e43d5890862.gif)


## Installation
1. In v5.16 and above, the easiest way to install a mMattermost plugin is by clicking on the "Settings" menu button above the channel list.  Select "Plugin Marketplace", search for "Webex" and click "Install"

2. Alternatively, download one of the releases from the [gitHub page](https://github.com/mattermost/mattermost-plugin-webex/releases)

3. Go to Settings --> Plugins --> Upload Plugin.  Select the file you downloaded, upload it to the server. In server 5.14+, plugins will automatically be distributed across an Enterprise cluster of Mattermost servers.

4. Go to settings --> PLugin Management and Enable the Webex Meeting Plugin 

## Configuration
1. Go to Settings --> Webex Plugin

2. Insert the Webex Meetings URL for your organization, it is often in the format of <mycompany>.my.webex.com

## Usage

If you type `/webex help` in any channel conversation you will be presented with your available options.


### Starting a Meeting

There are two primary ways of initiating a new Webex Meeting from within Mattermost:
  - Clicking the Webex Meeting Button at the top right of the channel
  - Typing in `/webex start` in a chat window
  

## Development

This plugin contains both a server and web app portion.

Use `make dist` to build distributions of the plugin that you can upload to a Mattermost server for testing.

Use `make check-style` to check the style for the whole plugin.

### Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within there, build the plugin like you would any other Go application.

### Web App

Inside the `/webapp` directory, you will find the JS and React files that make up the client-side of the plugin. Within there, modify files and components as necessary. Test your syntax by running `npm run build`.
