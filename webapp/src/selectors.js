import {getConfig} from 'mattermost-redux/selectors/entities/general';

export const getServerRoute = (state) => {
    const config = getConfig(state);

    let basePath = '';
    if (config && config.SiteURL) {
        try {
            basePath = new URL(config.SiteURL).pathname;

            if (basePath && basePath[basePath.length - 1] === '/') {
                basePath = basePath.substr(0, basePath.length - 1);
            }
        } catch (e) {
            basePath = '';
        }
    }

    return basePath;
};
