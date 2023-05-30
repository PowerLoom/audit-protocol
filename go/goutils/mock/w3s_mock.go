package mock

type W3SMock struct {
	UploadToW3sMock func(msg interface{}) (string, error)
}

func (m W3SMock) UploadToW3s(msg interface{}) (string, error) {
	return m.UploadToW3sMock(msg)
}
