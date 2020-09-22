package main

type mockStore struct {
	userInfo UserInfo
}

func (store mockStore) StoreUserInfo(mattermostUserID string, info UserInfo) error {
	return nil
}
func (store mockStore) LoadUserInfo(mattermostUserID string) (UserInfo, error) {
	return store.userInfo, nil
}
