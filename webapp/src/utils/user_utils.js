// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {getFullName} from 'mattermost-redux/utils/user_utils';
import {Preferences} from 'mattermost-redux/constants';

function nameDisplaySetting(config) {
    const key = `${Preferences.CATEGORY_DISPLAY_SETTINGS}--${Preferences.NAME_NAME_FORMAT}`;
    const prefs = config.entities.preferences.myPreferences;
    if (!(key in prefs)) {
        return 'username';
    }

    return prefs[key].value;
}

export function displayUsernameForUser(user, config) {
    if (user) {
        const nameFormat = nameDisplaySetting(config);
        let name = user.username;
        if (nameFormat === 'nickname_full_name' && user.nickname && user.nickname !== '') {
            name = user.nickname;
        } else if ((user.first_name || user.last_name) && (nameFormat === 'nickname_full_name' || nameFormat === 'full_name')) {
            name = getFullName(user);
        }

        return name;
    }

    return '';
}
