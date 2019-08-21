package main

type mockStore struct{}

func (store mockStore) StoreUserInfo(mattermostUserId string, info UserInfo) error {
	return nil
}
func (store mockStore) LoadUserInfo(mattermostUserId string) (UserInfo, error) {
	return UserInfo{RoomID: "myroom", Email: "myemail@host.com"}, nil
}
