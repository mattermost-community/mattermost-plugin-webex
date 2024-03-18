package main

type mockStore struct {
	userInfo UserInfo
}

func (store mockStore) StoreUserInfo(_ string, _ UserInfo) error {
	return nil
}
func (store mockStore) LoadUserInfo(_ string) (UserInfo, error) {
	return store.userInfo, nil
}
