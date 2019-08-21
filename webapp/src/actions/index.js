// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {PostTypes} from 'mattermost-redux/action_types';

import Client from '../client';

export function startMeeting(channelId) {
    return async (dispatch, getState) => {
        try {
            await Client.startMeeting(channelId, true);
        } catch (error) {
            if (error.response && error.response.statusCode === 400) {
                return {error};
            }
            let m = 'We could not verify your Mattermost account in Webex. Please ensure that your Mattermost email address matches your Webex email address.';
            if (error.response && error.response.text) {
                const e = JSON.parse(error.response.text);
                if (e && e.message) {
                    m += '\nWebex error: ' + e.message;
                }
            }
            const post = {
                id: 'webexPlugin' + Date.now(),
                create_at: Date.now(),
                update_at: 0,
                edit_at: 0,
                delete_at: 0,
                is_pinned: false,
                user_id: getState().entities.users.currentUserId,
                channel_id: channelId,
                root_id: '',
                parent_id: '',
                original_id: '',
                message: m,
                type: 'system_ephemeral',
                props: {},
                hashtags: '',
                pending_post_id: '',
            };

            dispatch({
                type: PostTypes.RECEIVED_NEW_POST,
                data: post,
                channelId,
            });

            return {error};
        }

        return {data: true};
    };
}
