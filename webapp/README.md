# README for Zoom

Add README information for the webapp portion of your plugin here.

## Install
1. Follow the instructions to add Mattermost as an integration: https://developer.cisco.com/docs/webex-meetings/#!integration/what-are-integrations
1. First, log into Cisco Devnet (the instruction site in step 1). 
1. Go to: https://developer.cisco.com/site/webex-integration and register Mattermost as an integration.
1. For the `Redirect URI` enter `https://<mattermost_url>/plugins/com.mattermost.webex/oauth` where `<mattermost_url>` is the hostname for your Mattermost server.
1. For `Scope`, select `read_all` and `modify_meetings`.
1. In the description put anything to help you remember this integration. Eg, `This is the integration to allow Mattermost users to start Webex meetings within Mattermost.` 
1. Copy the ClientID and ClientSecret and enter them into those fields in the Webex plugin's system console settings.

## Template

Everything you need to build the webapp portion of your plugin is present in this directory.

### index.js

This is the entry point for your webapp plugin. Includes initilization code to handle registering your plugin with the Mattermost web and desktop apps. Use this file for additional set up or initialization.

### package.json

The minimum required dependencies will be added by default. Use this file for additional dependecies and npm targets as needed. This should be familiar if you have experience with npm, if not, [please take some time to learn about npm](https://www.npmjs.com/).

### components

The meat of your plugin will be the React components in this directory. You can find different directories and files depending on the components you chose to override. The default props that each component has access to are already defined. Use the `index.js` containers to supply new props and actions to the components as needed. Also include any child components you may need to build in this directory.

### client

Any web utilities you need to build for accessing different servers are added here. If you only need to access the existing Mattermost REST API, please use [mattermost-redux](https://github.com/mattermost/mattermost-redux), which is already included as a dependency. There should be a short example file to help illustrate the usage.

### actions

Your functions that affect the state of your plugins are in this directory. We recommend following [the pattern used in mattermost-redux](https://github.com/mattermost/mattermost-redux/blob/master/src/actions/users.js#L1253).

### webpack.config.js

Webpack is used to bundle the modules of your webapp plugin. Changes are typically not required.
