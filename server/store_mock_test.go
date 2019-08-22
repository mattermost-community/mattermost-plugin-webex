package main

type mockStore struct {
	userInfo UserInfo
}

func (store mockStore) StoreUserInfo(mattermostUserId string, info UserInfo) error {
	return nil
}
func (store mockStore) LoadUserInfo(mattermostUserId string) (UserInfo, error) {
	return store.userInfo, nil
}
