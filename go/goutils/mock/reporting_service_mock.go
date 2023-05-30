package mock

import "audit-protocol/goutils/datamodel"

type ReportingServiceMock struct {
	ReportMock func(issueType datamodel.IssueType, projectID string, epochID string, extra map[string]interface{})
}

func (m ReportingServiceMock) Report(issueType datamodel.IssueType, projectID string, epochID string, extra map[string]interface{}) {
	m.ReportMock(issueType, projectID, epochID, extra)
}
