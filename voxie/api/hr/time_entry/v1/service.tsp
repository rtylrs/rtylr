import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.time_entry.v1;

/** TimeEntryService defines HR time entry and timesheet endpoints. */
@route("/hr/time-entry/v1")
interface TimeEntryService {
  /** Returns time entries with optional filters. */
  @get
  @route("/time-entries")
  listTimeEntries(...ListTimeEntriesRequest): ListTimeEntriesResponse;

  /** Creates a new time entry. */
  @post
  @route("/time-entries")
  createTimeEntry(...CreateTimeEntryRequest): CreateTimeEntryResponse;

  /** Returns timesheets with optional filters. */
  @get
  @route("/timesheets")
  listTimesheets(...ListTimesheetsRequest): ListTimesheetsResponse;

  /** Submits a timesheet for approval. */
  @post
  @route("/timesheets")
  submitTimesheet(...SubmitTimesheetRequest): SubmitTimesheetResponse;

  /** Approves a submitted timesheet. */
  @patch
  @route("/timesheets/{id}/approve")
  approveTimesheet(...ApproveTimesheetRequest): ApproveTimesheetResponse;

  /**
   * Records that one finance invoice bills the listed time entries, so
   * unbilled-billable queries stop offering hours that have been invoiced.
   */
  @patch
  @route("/time-entries/mark-invoiced")
  markTimeEntriesInvoiced(
    ...MarkTimeEntriesInvoicedRequest,
  ): MarkTimeEntriesInvoicedResponse;
}
