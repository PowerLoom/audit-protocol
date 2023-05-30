package mock

import "audit-protocol/goutils/datamodel"

type IPFSServiceMock struct {
	UploadSnapshotToIPFSMock func(snapshot *datamodel.PayloadCommitMessage) error
	GetSnapshotFromIPFSMock  func(snapshotCID string, outputPath string) error
	UnpinMock                func(cid string) error
}

func (m IPFSServiceMock) UploadSnapshotToIPFS(snapshot *datamodel.PayloadCommitMessage) error {
	return m.UploadSnapshotToIPFSMock(snapshot)
}

func (m IPFSServiceMock) GetSnapshotFromIPFS(snapshotCID string, outputPath string) error {
	return m.GetSnapshotFromIPFSMock(snapshotCID, outputPath)
}

func (m IPFSServiceMock) Unpin(cid string) error {
	return m.UnpinMock(cid)
}
