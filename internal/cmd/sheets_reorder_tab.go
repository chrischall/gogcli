package cmd

import (
	"context"
	"os"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"

	"github.com/steipete/gogcli/internal/outfmt"
	"github.com/steipete/gogcli/internal/ui"
)

// SheetsReorderTabCmd moves a tab to a specific 0-based position in the
// spreadsheet via spreadsheets.batchUpdate -> updateSheetProperties with field
// mask `index`. Existing tab management (add/rename/delete) does not expose
// this — see #603.
type SheetsReorderTabCmd struct {
	SpreadsheetID string `arg:"" name:"spreadsheetId" help:"Spreadsheet ID"`
	Tab           string `name:"tab" required:"" help:"Target tab by name or numeric sheet ID (see sheets metadata)"`
	To            *int64 `name:"to" required:"" help:"Destination 0-based tab index"`
}

func (c *SheetsReorderTabCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)

	spreadsheetID := normalizeGoogleID(strings.TrimSpace(c.SpreadsheetID))
	tab := strings.TrimSpace(c.Tab)
	if spreadsheetID == "" {
		return usage("empty spreadsheetId")
	}
	if tab == "" {
		return usage("--tab is required")
	}
	if c.To == nil {
		return usage("--to is required")
	}
	if *c.To < 0 {
		return usage("--to must be >= 0")
	}

	if err := dryRunExit(ctx, flags, "sheets.reorder-tab", map[string]any{
		"spreadsheet_id": spreadsheetID,
		"tab":            tab,
		"to":             *c.To,
	}); err != nil {
		return err
	}

	account, err := requireAccount(flags)
	if err != nil {
		return err
	}

	svc, err := newSheetsService(ctx, account)
	if err != nil {
		return err
	}

	sheetID, resolvedTitle, err := resolveSheetTabID(ctx, svc, spreadsheetID, tab)
	if err != nil {
		return err
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId:         sheetID,
						Index:           *c.To,
						ForceSendFields: []string{"Index"},
					},
					Fields: "index",
				},
			},
		},
	}

	if _, err := svc.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do(); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		payload := map[string]any{
			"spreadsheetId": spreadsheetID,
			"sheetId":       sheetID,
			"index":         *c.To,
		}
		if resolvedTitle != "" {
			payload["title"] = resolvedTitle
		}
		return outfmt.WriteJSON(ctx, os.Stdout, payload)
	}

	if resolvedTitle != "" {
		u.Out().Linef("Moved tab %q (sheetId %d) to index %d in spreadsheet %s", resolvedTitle, sheetID, *c.To, spreadsheetID)
	} else {
		u.Out().Linef("Moved sheetId %d to index %d in spreadsheet %s", sheetID, *c.To, spreadsheetID)
	}
	return nil
}

// resolveSheetTabID accepts either a tab title or a numeric sheet ID and
// returns (sheetID, title). When the caller passed a numeric ID, title is
// empty unless we happen to find a matching tab while listing IDs.
func resolveSheetTabID(ctx context.Context, svc *sheets.Service, spreadsheetID, tab string) (int64, string, error) {
	if id, err := strconv.ParseInt(tab, 10, 64); err == nil {
		// Numeric — try to enrich with title for nicer output, but accept it
		// even if the GET fails or the tab isn't found in the map.
		if titles, mapErr := fetchSheetIDMap(ctx, svc, spreadsheetID); mapErr == nil {
			for title, sheetID := range titles {
				if sheetID == id {
					return id, title, nil
				}
			}
		}
		return id, "", nil
	}

	sheetIDs, err := fetchSheetIDMap(ctx, svc, spreadsheetID)
	if err != nil {
		return 0, "", err
	}
	sheetID, ok := sheetIDs[tab]
	if !ok {
		return 0, "", usagef("unknown tab %q", tab)
	}
	return sheetID, tab, nil
}
