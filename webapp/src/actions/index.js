// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {PostTypes} from 'mattermost-redux/action_types';

import Client from '../client';

export function startMeeting(channelId) {
    return async (dispatch, getState) => {
        try {
            await Client.startMeeting(channelId, true);
        } catch (error) {
            if (error.response && error.response.text) {
                let m = 'We could not start a meeting.';
                m += '\nWebex error: ' + error.response.text;
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
            }

            return {error};
        }

        return {data: true};
    };
}
