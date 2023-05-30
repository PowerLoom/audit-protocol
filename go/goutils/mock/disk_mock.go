package mock

type DiskMock struct {
	ReadMock  func(filepath string) ([]byte, error)
	WriteMock func(filepath string, data []byte) error
}

func (m DiskMock) Read(filepath string) ([]byte, error) {
	return m.ReadMock(filepath)
}

func (m DiskMock) Write(filepath string, data []byte) error {
	return m.WriteMock(filepath, data)
}
