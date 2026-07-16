import {getConfig} from 'mattermost-redux/selectors/entities/general';

import {getServerRoute} from './selectors';

jest.mock('mattermost-redux/selectors/entities/general', () => ({
    getConfig: jest.fn(),
}));

describe('getServerRoute', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('returns empty string when config is missing', () => {
        getConfig.mockReturnValue(null);
        expect(getServerRoute({})).toBe('');
    });

    test('returns empty string when SiteURL is missing', () => {
        getConfig.mockReturnValue({});
        expect(getServerRoute({})).toBe('');
    });

    test('returns empty string when SiteURL has no subpath', () => {
        getConfig.mockReturnValue({SiteURL: 'https://example.com'});
        expect(getServerRoute({})).toBe('');
    });

    test('returns pathname when SiteURL has a subpath', () => {
        getConfig.mockReturnValue({SiteURL: 'https://example.com/mattermost'});
        expect(getServerRoute({})).toBe('/mattermost');
    });

    test('strips trailing slash from subpath', () => {
        getConfig.mockReturnValue({SiteURL: 'https://example.com/mattermost/'});
        expect(getServerRoute({})).toBe('/mattermost');
    });

    test('returns empty string when SiteURL is a bare root URL with trailing slash', () => {
        getConfig.mockReturnValue({SiteURL: 'https://example.com/'});
        expect(getServerRoute({})).toBe('');
    });

    test('returns empty string when SiteURL is malformed', () => {
        getConfig.mockReturnValue({SiteURL: 'not a valid url'});
        expect(getServerRoute({})).toBe('');
    });
});
