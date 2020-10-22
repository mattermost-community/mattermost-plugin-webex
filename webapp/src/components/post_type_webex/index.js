// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';

import {getBool} from 'mattermost-redux/selectors/entities/preferences';

import {displayUsernameForUser} from '../../utils/user_utils';

import PostTypeWebex from './post_type_webex.jsx';

function mapStateToProps(state, ownProps) {
    const post = ownProps.post || {};
    const user = state.entities.users.profiles[post.props.starting_user_id] || {};

    return {
        ...ownProps,
        fromBot: ownProps.post.props.from_bot,
        creatorName: displayUsernameForUser(user, state),
        useMilitaryTime: getBool(state, 'display_settings', 'use_military_time', false),
    };
}

function mapDispatchToProps(dispatch) {
    return {
        actions: bindActionCreators({
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(PostTypeWebex);
