/**
 * @jest-environment jsdom
 */

/* eslint-disable max-nested-callbacks */

jest.mock('./components/icon.jsx', () => ({__esModule: true, default: () => null}));
jest.mock('./components/post_type_webex', () => ({__esModule: true, default: () => null}));

jest.mock('./actions', () => ({
    startMeeting: jest.fn(() => jest.fn()),
}));

jest.mock('./client', () => ({
    __esModule: true,
    default: {
        setServerRoute: jest.fn(),
    },
}));

jest.mock('mattermost-redux/selectors/entities/general', () => ({
    getConfig: jest.fn(() => ({SiteURL: 'https://example.com'})),
}));

jest.mock('./selectors', () => ({
    getServerRoute: jest.fn(() => ''),
}));

describe('Plugin initialization', () => {
    let mockRegistry;
    let mockStore;

    beforeEach(() => {
        jest.clearAllMocks();

        mockRegistry = {
            registerChannelHeaderButtonAction: jest.fn(),
            registerAppBarComponent: jest.fn(),
            registerPostTypeComponent: jest.fn(),
        };
        mockStore = {
            getState: jest.fn(() => ({})),
            dispatch: jest.fn(),
        };

        // Mirror modern Mattermost: window.registerPlugin invokes initialize()
        // synchronously during registration. If any module-scoped identifier
        // used inside initialize() is still in the temporal dead zone at this
        // point (e.g. a const arrow declared after registerPlugin), requiring
        // ./index will throw a ReferenceError.
        window.registerPlugin = jest.fn((_pluginId, plugin) => {
            plugin.initialize(mockRegistry, mockStore);
        });
    });

    test('module evaluation with synchronous initialize() does not throw', () => {
        expect(() => {
            jest.isolateModules(() => {
                require('./index'); // eslint-disable-line global-require
            });
        }).not.toThrow();
    });

    test('registers channel header button, app bar icon, and post type component', () => {
        jest.isolateModules(() => {
            require('./index'); // eslint-disable-line global-require
        });

        expect(mockRegistry.registerChannelHeaderButtonAction).toHaveBeenCalledTimes(1);
        expect(mockRegistry.registerAppBarComponent).toHaveBeenCalledTimes(1);
        expect(mockRegistry.registerPostTypeComponent).toHaveBeenCalledWith(
            'custom_webex',
            expect.anything(),
        );
    });
});
