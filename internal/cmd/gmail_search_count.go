package cmd

import (
	"context"
	"fmt"

	"google.golang.org/api/gmail/v1"

	"github.com/openclaw/gogcli/internal/ui"
)

// One maximal page is enough to count all but the broadest result sets, and
// 500 is the ceiling Gmail's list endpoints accept.
const gmailCountProbePageSize = 500

// gmailMatchCount is how large a result set really is.
//
// Exact means the probe reached the end of the set, so Value is the total.
// Otherwise the probe filled its page with more behind it and Value is a lower
// bound — reported as such rather than rounded into a total nobody can trust.
type gmailMatchCount struct {
	Value int64
	Exact bool
}

// apply writes the count into a JSON payload under the name that matches its
// certainty, so a consumer never has to guess whether a number is a total.
func (c gmailMatchCount) apply(payload map[string]any) {
	if c.Exact {
		payload["totalMatches"] = c.Value
		return
	}
	payload["totalMatchesAtLeast"] = c.Value
}

// Deliberately NOT Gmail's resultSizeEstimate, which is the obvious source and
// saturates: measured against a live mailbox it returned exactly 201 for every
// non-empty query — from:freshbooks.com (21 real matches), from:housecallpro.com
// (6), from:thumbtack.com newer_than:30d (3) — and 0 for a query with none. It
// is a has-results boolean wearing a number's clothes, and it does not vary
// with maxResults either. Emitting it would let a caller report "3 of ~201"
// when the truth is 3 of 6. See openclaw/gogcli#983.
//
// Counting bare ids costs one extra list call and returns a response of ids
// alone, and it is exact whenever the set fits a single page — the common case,
// and always the case for the narrow queries where a wrong count does the most
// damage.
func countGmailThreadMatches(ctx context.Context, svc *gmail.Service, query string) (gmailMatchCount, error) {
	opts := newGmailSearchRequestOptions(query, gmailCountProbePageSize, "")
	resp, err := applyGmailThreadListOptions(svc.Users.Threads.List("me"), opts).
		Fields("threads/id,nextPageToken").
		Context(ctx).
		Do()
	if err != nil {
		return gmailMatchCount{}, err
	}
	return gmailMatchCount{Value: int64(len(resp.Threads)), Exact: resp.NextPageToken == ""}, nil
}

func countGmailMessageMatches(ctx context.Context, svc *gmail.Service, query string) (gmailMatchCount, error) {
	opts := newGmailSearchRequestOptions(query, gmailCountProbePageSize, "")
	resp, err := applyGmailMessageListOptions(svc.Users.Messages.List("me"), opts).
		Fields("messages/id,nextPageToken").
		Context(ctx).
		Do()
	if err != nil {
		return gmailMatchCount{}, err
	}
	return gmailMatchCount{Value: int64(len(resp.Messages)), Exact: resp.NextPageToken == ""}, nil
}

// The human-facing form of the same fact, on stderr so the table on stdout
// stays parseable. Says outright when the page is not the whole set: a caller
// reading only the visible rows is exactly how a partial result gets reported
// as an absence.
func printGmailMatchCount(u *ui.UI, shown int, count gmailMatchCount) {
	if count.Exact {
		if int64(shown) < count.Value {
			u.Err().Println(fmt.Sprintf("Showing %d of %d matches.", shown, count.Value))
			return
		}
		u.Err().Println(fmt.Sprintf("%d matches.", count.Value))
		return
	}
	u.Err().Println(fmt.Sprintf("Showing %d of at least %d matches.", shown, count.Value))
}
