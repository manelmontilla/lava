// Copyright 2023 Adevinta

package engine

import (
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/adevinta/vulcan-agent/storage"
	report "github.com/adevinta/vulcan-report"
)

// reportStore stores the reports generated by the Vulcan agent in
// memory. It implements [storage.Store].
type reportStore struct {
	mu      sync.Mutex
	reports map[string]report.Report
}

var _ storage.Store = &reportStore{}

// UploadCheckData decodes the provided content and stores it in
// memory indexed by checkID. If kind is "reports", it decodes content
// as [report.Report]. If kind is "logs", the data is ignored.
func (rs *reportStore) UploadCheckData(checkID, kind string, startedAt time.Time, content []byte) (link string, err error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	logger := slog.With("checkID", checkID)

	if rs.reports == nil {
		rs.reports = make(map[string]report.Report)
	}

	switch kind {
	case "reports":
		logger.Debug("received reports from check", "content", fmt.Sprintf("%#q", content))

		var r report.Report
		if err := r.UnmarshalJSONTimeAsString(content); err != nil {
			return "", fmt.Errorf("decode content: %w", err)
		}
		rs.reports[checkID] = r
	case "logs":
		logger.Debug("received logs from check", "content", fmt.Sprintf("%#q", content))
	default:
		return "", fmt.Errorf("unknown data kind: %v", kind)
	}
	return "", nil
}

// Summary returns a human-readable summary per report.
func (rs *reportStore) Summary() []string {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var sums []string
	for _, r := range rs.reports {
		s := fmt.Sprintf("checktype=%v target=%v start=%v status=%v", r.ChecktypeName, r.Target, r.StartTime, r.Status)
		sums = append(sums, s)
	}
	return sums
}

// Reports returns the stored reports.
func (rs *reportStore) Reports() map[string]report.Report {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return maps.Clone(rs.reports)
}
