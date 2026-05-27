// Package sheets provides a thin Google Sheets v4 integration for appending
// rows to a spreadsheet. It is used by the websitelead module to log every
// incoming lead to a Google Sheet so Abed can track pipeline without a CRM.
//
// Setup (one-time):
//  1. Create a Google Sheet and note its ID (the long string in the URL).
//  2. In Google Cloud Console, create a Service Account and download the JSON key.
//  3. Share the sheet with the service account's email address (Editor access).
//  4. Base64-encode the JSON key: `base64 < service-account.json`
//  5. Set the following env vars on the deployment platform:
//     GOOGLE_SHEETS_CREDENTIALS   — base64-encoded service account JSON
//     GOOGLE_SHEETS_LEAD_SHEET_ID — the spreadsheet ID
//
// When either env var is absent, AppendLeadRow is a no-op and logs a warning.
package sheets

import (
	"context"
	"encoding/base64"
	"os"
	"sync"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sheetsv4 "google.golang.org/api/sheets/v4"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

const leadsSheetRange = "Leads!A:J"

var (
	once    sync.Once
	svc     *sheetsv4.Service
	sheetID string
)

func Initialize(ctx context.Context) {
	once.Do(func() {
		if config.GlobalConfig != nil {
			sheetID = config.GlobalConfig.GoogleSheetsLeadSheetID
		} else {
			sheetID = os.Getenv("GOOGLE_SHEETS_LEAD_SHEET_ID")
		}
		if sheetID == "" {
			logger.Global().Debug("GOOGLE_SHEETS_LEAD_SHEET_ID not set — Sheets logging disabled")
			return
		}

		var encoded string
		if config.GlobalConfig != nil {
			encoded = config.GlobalConfig.GoogleSheetsCredentials
		} else {
			encoded = os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
		}
		if encoded == "" {
			logger.Global().Debug("GOOGLE_SHEETS_CREDENTIALS not set — Sheets logging disabled")
			return
		}

		jsonKey, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			logger.Global().Error("Failed to base64-decode GOOGLE_SHEETS_CREDENTIALS", logger.Error(err))
			return
		}

		creds, err := google.CredentialsFromJSONWithType(ctx, jsonKey, google.ServiceAccount, sheetsv4.SpreadsheetsScope)
		if err != nil {
			logger.Global().Error("Failed to parse Google service account credentials", logger.Error(err))
			return
		}

		svc, err = sheetsv4.NewService(ctx, option.WithCredentials(creds))
		if err != nil {
			logger.Global().Error("Failed to create Google Sheets client", logger.Error(err))
			return
		}

		logger.Global().Info("Google Sheets client initialized",
			logger.String("sheet_id", sheetID),
		)
	})
}

// IsEnabled returns true when the Sheets client has been successfully initialized.
func IsEnabled() bool { return svc != nil && sheetID != "" }

// AppendLeadRow appends one row to the "Leads" sheet tab.
// Columns: ID | Timestamp | Name | Email | Phone | Company | Service | Budget | Source | Project
// Errors are logged but never propagated — Sheets logging is best-effort.
func AppendLeadRow(ctx context.Context, values []interface{}) {
	if !IsEnabled() {
		return
	}

	row := &sheetsv4.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := svc.Spreadsheets.Values.
		Append(sheetID, leadsSheetRange, row).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to append lead row to Google Sheets",
			logger.Error(err),
			logger.String("sheet_id", sheetID),
		)
	}
}
