import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.sales.v1;

/**
 * Contract is a sales agreement: what a customer bought, what it costs and
 * how often it is billed, its term, and who signed. Winning a deal drafts one.
 *
 * A contract is "signed" when signed_at is present. Unsigned drafts exist so
 * sales can stage terms before signature, and callers must treat the two
 * states differently.
 */
model Contract {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /**
   * lead_id is the lead the contract's deal came from; absent for deals
   * opened directly and contracts recorded without a deal.
   */
  @encode(string)
  lead_id?: int64;

  /**
   * deal_id is the won deal this contract records; absent for contracts
   * recorded without one.
   */
  @encode(string)
  deal_id?: int64;

  /** plan is what was sold, as free text (e.g. "Annual maintenance"). */
  plan: string;

  /**
   * mrr is the monthly equivalent of value as a decimal string, because
   * money crosses the wire as text to avoid float drift: value when billed
   * monthly, a twelfth when yearly, "0.00" for a one-time sale. Currency is
   * the business's reporting currency.
   */
  mrr: string;

  /** seats is how many users the contract covers. */
  seats: int32;

  /** value is what is charged per billing period (decimal string). */
  value: string;

  /** billing is one_time, monthly or yearly. */
  billing: string;

  /** start_date is when the contract's term starts (YYYY-MM-DD). */
  start_date?: string;

  /** end_date is when the term ends (YYYY-MM-DD); absent for open-ended. */
  end_date?: string;

  /**
   * customer_id is the CRM customer the contract belongs to, linked from
   * the lead's contact.
   */
  @encode(string)
  customer_id?: int64;

  /**
   * signed_at is when the customer signed (ISO-8601); absent on unsigned
   * drafts. Presence of this field is what makes a contract billable.
   */
  signed_at?: string;

  /** signed_by_name is the customer-side signer as written at signature time. */
  signed_by_name: string;

  /** notes are free-text sales notes; not for invoice prefill. */
  notes: string;

  /** created_at is when the record was created. */
  @encode(string)
  created_at: int64;

  /** updated_at is when the record was last updated. */
  @encode(string)
  updated_at: int64;
}

/**
 * ListContractsRequest selects contracts for this business. Filters combine
 * with AND. The response is always a page: follow offset until total is
 * reached.
 */
model ListContractsRequest {
  /**
   * signed filters to contracts whose signing is complete when true, or
   * incomplete when false; absent to include both.
   */
  @query("signed")
  signed?: boolean;

  /** Maximum number of records to return, capped by the server. */
  @query("limit")
  limit: int32;

  /** Number of records to skip for pagination. */
  @query("offset")
  offset: int32;
}

/**
 * ListContractsResponse returns one page of contracts with paging echoes,
 * matching the rest of this module's list surfaces.
 */
model ListContractsResponse {
  /** contracts is the page of contracts matching the filters. */
  contracts: Contract[];

  /** total is the number of contracts across all pages. */
  total: int64;

  /** limit echoes the applied page size. */
  limit: int32;

  /** offset echoes the applied skip. */
  offset: int32;
}

/** GetContractRequest selects one contract by id. */
model GetContractRequest {
  /** id is the contract id path parameter. */
  @encode(string)
  id: int64;
}

/**
 * GetContractResponse returns the requested contract, or a 404 error when it
 * does not exist in the caller's business.
 */
model GetContractResponse {
  /** contract is the full record. */
  contract: Contract;
}

/**
 * CreateContractRequest drafts a contract, usually for a won deal. Signing
 * can happen now or later via update.
 */
model CreateContractRequest {
  /** deal_id links the contract to the deal it records. */
  @minValueExclusive(0)
  @encode(string)
  deal_id?: int64;

  /**
   * customer_id is the CRM customer the contract belongs to; the deal's
   * customer when absent.
   */
  @minValueExclusive(0)
  @encode(string)
  customer_id?: int64;

  /** plan is what was bought, as CRM displays it. */
  @Onkt.required
  plan: string;

  /** seats is how many users the contract covers. */
  @minValue(0)
  seats: int32;

  /** value is what is charged per billing period (decimal string). */
  @Onkt.required
  value: string;

  /** billing is one_time, monthly or yearly; monthly when absent. */
  billing?: string;

  /** start_date is when the term starts (YYYY-MM-DD). */
  start_date?: string;

  /** end_date is when the term ends (YYYY-MM-DD). */
  end_date?: string;

  /**
   * signed_at completes the signature at creation time; absent means an
   * unsigned draft.
   */
  signed_at?: string;

  /** signed_by_name names the customer-side signer. */
  signed_by_name: string;

  /** notes are free-text sales notes. */
  notes: string;
}

/** CreateContractResponse returns the newly staged contract. */
model CreateContractResponse {
  /** contract is the record as stored. */
  contract: Contract;
}

/**
 * UpdateContractRequest replaces a contract's terms. The customer is kept
 * when the body names none.
 */
model UpdateContractRequest {
  /** id is the contract id path parameter. */
  @encode(string)
  id: int64;

  /** body carries the amendable fields bound from the HTTP payload. */
  @body
  body: UpdateContractBody;
}

/** UpdateContractBody is the amendable surface of a contract. */
model UpdateContractBody {
  /** plan replaces what the contract says was bought. */
  plan: string;

  /** seats replaces the covered user count. */
  seats: int32;

  /** value replaces what is charged per billing period (decimal string). */
  value: string;

  /** billing replaces the billing period: one_time, monthly or yearly. */
  billing?: string;

  /** start_date replaces when the term starts (YYYY-MM-DD). */
  start_date?: string;

  /** end_date replaces when the term ends (YYYY-MM-DD). */
  end_date?: string;

  /**
   * signed_at records the signature (YYYY-MM-DD); setting it on an unsigned
   * draft makes the contract billable, and leaving it absent unsigns.
   */
  signed_at?: string;

  /** signed_by_name replaces the customer-side signer. */
  signed_by_name: string;

  /** notes replace the free-text sales notes. */
  notes: string;
}

/** UpdateContractResponse returns the amended contract. */
model UpdateContractResponse {
  /** contract is the record after amendment. */
  contract: Contract;
}

// ---------------------------------------------------------------------------
// Leads: prospects to qualify (docs/crm-model.md in services/onion).
// Money crosses the wire as decimal strings. Dates are YYYY-MM-DD; instants
// are Unix seconds.
// ---------------------------------------------------------------------------
/** Lead is a prospect to qualify: a person and the company they work for. */
model Lead {
  /** id is the lead's identifier. */
  @encode(string)
  id: int64;

  /** first_name is the person's first name. */
  first_name: string;

  /** last_name is the person's last name. */
  last_name: string;

  /** job_title is the person's role at the company. */
  job_title: string;

  /** email is the person's email address, lowercased. */
  email: string;

  /** phone is the person's phone number as entered. */
  phone: string;

  /** company_name is the company the person works for. */
  company_name: string;

  /** website is the company's website. */
  website: string;

  /** linkedin is the person's LinkedIn profile. */
  linkedin: string;

  /** industry is the company's industry. */
  industry: string;

  /** size is the company's size band. */
  size: string;

  /** city is where the lead is. */
  city: string;

  /** country is where the lead is. */
  country: string;

  /** source is where the lead came from, e.g. "LinkedIn" or a list name. */
  source: string;

  /** status is new, attempting, connected, nurture, qualified or disqualified. */
  status: string;

  /** disqualify_reason says why a disqualified lead was dropped. */
  disqualify_reason: string;

  /** assigned_to is the rep working the lead. */
  @encode(string)
  assigned_to?: int64;

  /** value is the estimated deal value (decimal string), when known. */
  value?: string;

  /** billing is how often value would be charged: one_time, monthly or yearly. */
  billing: string;

  /** notes are free-text notes. */
  notes: string;

  /** attempt_count is how many calls have been logged. */
  attempt_count: int32;

  /** last_activity_at is when the lead was last worked. */
  @encode(string)
  last_activity_at?: int64;

  /** next_follow_up is the earliest open task's due date. */
  next_follow_up?: string;

  /** customer_id is the CRM customer the lead became on conversion. */
  @encode(string)
  customer_id?: int64;

  /** company_id is the company record created or matched on conversion. */
  @encode(string)
  company_id?: int64;

  /** deal_id is the deal opened on conversion. */
  @encode(string)
  deal_id?: int64;

  /** converted_at is when the lead was qualified. */
  @encode(string)
  converted_at?: int64;

  /** import_batch_id is the import that created the lead. */
  @encode(string)
  import_batch_id?: int64;

  /** created_at is when the lead was created. */
  @encode(string)
  created_at: int64;

  /** updated_at is when the lead was last changed. */
  @encode(string)
  updated_at: int64;
}

/** LeadFields are a lead's editable fields. */
model LeadFields {
  /** first_name is the person's first name. */
  first_name: string;

  /** last_name is the person's last name. */
  last_name: string;

  /** job_title is the person's role. */
  job_title: string;

  /** email is the person's email address. */
  email: string;

  /** phone is the person's phone number. */
  phone: string;

  /** company_name is the company; a lead needs it or a person's name. */
  company_name: string;

  /** website is the company's website. */
  website: string;

  /** linkedin is the person's LinkedIn profile. */
  linkedin: string;

  /** industry is the company's industry. */
  industry: string;

  /** size is the company's size band. */
  size: string;

  /** city is where the lead is. */
  city: string;

  /** country is where the lead is. */
  country: string;

  /** source is where the lead came from. */
  source: string;

  /** assigned_to is the rep to work the lead; absent leaves it unassigned. */
  @encode(string)
  assigned_to?: int64;

  /** value is the estimated deal value (decimal string). */
  value?: string;

  /** billing is one_time, monthly or yearly; one_time when absent. */
  billing: string;

  /** notes are free-text notes. */
  notes: string;
}

/** ListLeadsRequest selects one page of leads. Filters combine with AND. */
model ListLeadsRequest {
  /** statuses is a comma-separated list of statuses; all when empty. */
  @query("statuses")
  statuses: string;

  /**
   * reps is a comma-separated list of rep user ids; 0 selects unassigned
   * leads.
   */
  @query("reps")
  reps: string;

  /** sources is a JSON array of sources to include. */
  @query("sources")
  sources: string;

  /** industries is a JSON array of industries to include. */
  @query("industries")
  industries: string;

  /** sizes is a JSON array of company sizes to include. */
  @query("sizes")
  sizes: string;

  /** countries is a JSON array of countries to include. */
  @query("countries")
  countries: string;

  /** cities is a JSON array of cities to include. */
  @query("cities")
  cities: string;

  /** billing is a comma-separated list of billing periods. */
  @query("billing")
  billing: string;

  /** import_id filters to one import's leads. */
  @query("import_id")
  @encode(string)
  import_id?: int64;

  /**
   * follow_up is overdue, today, due (overdue or today), upcoming,
   * scheduled (any date) or none.
   */
  @query("follow_up")
  follow_up: string;

  /**
   * has_phone is yes or no; any when empty. The other has_* fields work the
   * same way.
   */
  @query("has_phone")
  has_phone: string;

  @query("has_email")
  has_email: string;

  @query("has_website")
  has_website: string;

  @query("has_linkedin")
  has_linkedin: string;

  @query("has_company")
  has_company: string;

  /** has_value is whether an estimated value is set. */
  @query("has_value")
  has_value: string;

  /** contacted is yes (worked at least once) or no (never worked). */
  @query("contacted")
  contacted: string;

  /** attempts_min and attempts_max bound the call attempts, inclusive. */
  @query("attempts_min")
  attempts_min?: int32;

  @query("attempts_max")
  attempts_max?: int32;

  /** value_min and value_max bound the estimated value (decimal strings). */
  @query("value_min")
  value_min: string;

  @query("value_max")
  value_max: string;

  /**
   * created_from and created_to bound the date added (YYYY-MM-DD, the
   * caller's days, inclusive).
   */
  @query("created_from")
  created_from: string;

  @query("created_to")
  created_to: string;

  /** active_from and active_to bound the last activity date the same way. */
  @query("active_from")
  active_from: string;

  @query("active_to")
  active_to: string;

  /** search matches names, company, email, job title and phone. */
  @query("search")
  search: string;

  /** sort is newest, follow_up, recent_activity or company. */
  @query("sort")
  sort: string;

  /** cursor continues from a previous page's next_cursor. */
  @query("cursor")
  cursor: string;

  /** limit is the page size, at most 200. */
  @query("limit")
  limit: int32;

  /** today is the caller's date, which follow-up windows are relative to. */
  @query("today")
  today: string;

  /**
   * utc_offset_minutes is the caller's offset from UTC, so date bounds start
   * at the caller's midnight.
   */
  @query("utc_offset_minutes")
  utc_offset_minutes: int32;
}

/** ListLeadsResponse is one page of leads. */
model ListLeadsResponse {
  /** leads is the page. */
  leads: Lead[];

  /** next_cursor continues the list; empty on the last page. */
  next_cursor: string;

  /** total is how many leads match the filters. */
  total: int64;
}

/** GetLeadStatsRequest selects whose prospecting to summarise. */
model GetLeadStatsRequest {
  /**
   * reps is a comma-separated list of rep user ids; 0 selects unassigned
   * leads.
   */
  @query("reps")
  reps: string;

  /** sources is a JSON array of sources to include. */
  @query("sources")
  sources: string;

  /** industries is a JSON array of industries to include. */
  @query("industries")
  industries: string;

  /** sizes is a JSON array of company sizes to include. */
  @query("sizes")
  sizes: string;

  /** countries is a JSON array of countries to include. */
  @query("countries")
  countries: string;

  /** cities is a JSON array of cities to include. */
  @query("cities")
  cities: string;

  /** billing is a comma-separated list of billing periods. */
  @query("billing")
  billing: string;

  /** import_id filters to one import's leads. */
  @query("import_id")
  @encode(string)
  import_id?: int64;

  /**
   * follow_up is overdue, today, due (overdue or today), upcoming,
   * scheduled (any date) or none.
   */
  @query("follow_up")
  follow_up: string;

  /**
   * has_phone is yes or no; any when empty. The other has_* fields work the
   * same way.
   */
  @query("has_phone")
  has_phone: string;

  @query("has_email")
  has_email: string;

  @query("has_website")
  has_website: string;

  @query("has_linkedin")
  has_linkedin: string;

  @query("has_company")
  has_company: string;

  /** has_value is whether an estimated value is set. */
  @query("has_value")
  has_value: string;

  /** contacted is yes (worked at least once) or no (never worked). */
  @query("contacted")
  contacted: string;

  /** attempts_min and attempts_max bound the call attempts, inclusive. */
  @query("attempts_min")
  attempts_min?: int32;

  @query("attempts_max")
  attempts_max?: int32;

  /** value_min and value_max bound the estimated value (decimal strings). */
  @query("value_min")
  value_min: string;

  @query("value_max")
  value_max: string;

  /**
   * created_from and created_to bound the date added (YYYY-MM-DD, the
   * caller's days, inclusive).
   */
  @query("created_from")
  created_from: string;

  @query("created_to")
  created_to: string;

  /** active_from and active_to bound the last activity date the same way. */
  @query("active_from")
  active_from: string;

  @query("active_to")
  active_to: string;

  /** search matches names, company, email, job title and phone. */
  @query("search")
  search: string;

  /** today is the caller's date. */
  @query("today")
  today: string;

  /**
   * utc_offset_minutes is the caller's offset from UTC, so "today" counts
   * start at the caller's midnight.
   */
  @query("utc_offset_minutes")
  utc_offset_minutes: int32;
}

/** LeadStatusCount is how many leads have one status. */
model LeadStatusCount {
  /** status is the lead status. */
  status: string;

  /** count is how many leads have it. */
  count: int64;
}

/** GetLeadStatsResponse summarises prospecting. */
model GetLeadStatsResponse {
  /** by_status counts leads per status. */
  by_status: LeadStatusCount[];

  /** follow_ups_overdue counts worked leads with an overdue follow-up. */
  follow_ups_overdue: int64;

  /** follow_ups_today counts worked leads with a follow-up due today. */
  follow_ups_today: int64;

  /** calls_today counts calls logged since the caller's midnight. */
  calls_today: int64;

  /** connected_today counts those calls that reached the person. */
  connected_today: int64;

  /** converted_this_month counts leads qualified this calendar month. */
  converted_this_month: int64;
}

/** ListLeadFacetsRequest takes no parameters. */
model ListLeadFacetsRequest {}

/** ListLeadFacetsResponse lists the distinct values leads have, for filters. */
model ListLeadFacetsResponse {
  /** sources are the distinct non-empty sources. */
  sources: string[];

  /** industries are the distinct non-empty industries. */
  industries: string[];

  /** sizes are the distinct non-empty company sizes. */
  sizes: string[];

  /** countries are the distinct non-empty countries. */
  countries: string[];

  /** cities are the distinct non-empty cities. */
  cities: string[];
}

/** GetLeadRequest selects one lead. */
model GetLeadRequest {
  /** id is the lead id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** GetLeadResponse returns one lead. */
model GetLeadResponse {
  /** lead is the record. */
  lead: Lead;
}

/** CreateLeadRequest adds a lead as new. */
model CreateLeadRequest {
  /** lead holds the fields. */
  lead: LeadFields;
}

/** CreateLeadResponse returns the created lead. */
model CreateLeadResponse {
  /** lead is the record as stored. */
  lead: Lead;
}

/** UpdateLeadRequest replaces a lead's fields and status. */
model UpdateLeadRequest {
  /** id is the lead id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new values. */
  @body
  body: UpdateLeadBody;
}

/** UpdateLeadBody is a full edit of a lead. */
model UpdateLeadBody {
  /** lead holds the fields. */
  lead: LeadFields;

  /** status is the new status; qualifying happens through conversion. */
  status: string;

  /** disqualify_reason is required when disqualifying. */
  disqualify_reason: string;
}

/** UpdateLeadResponse returns the edited lead. */
model UpdateLeadResponse {
  /** lead is the record as stored. */
  lead: Lead;
}

/** DeleteLeadRequest removes an unconverted lead. */
model DeleteLeadRequest {
  /** id is the lead id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** DeleteLeadResponse confirms the removal. */
model DeleteLeadResponse {
  /** deleted is true once the lead is gone. */
  deleted: boolean;
}

/** UpdateLeadsRequest assigns or re-statuses many leads at once. */
model UpdateLeadsRequest {
  /** ids are the leads to change, at most 1000. */
  @minItems(1)
  @Onkt.encode("string")
  ids: int64[];

  /** assign_to gives the leads to a rep. */
  @encode(string)
  assign_to?: int64;

  /** unassign takes the leads off whoever has them. */
  unassign: boolean;

  /** status sets the leads' status instead. */
  status: string;

  /** disqualify_reason is required when disqualifying. */
  disqualify_reason: string;
}

/** UpdateLeadsResponse reports how many leads changed. */
model UpdateLeadsResponse {
  /** updated is how many leads changed. */
  updated: int64;
}

/** DeleteLeadsRequest removes many unconverted leads. */
model DeleteLeadsRequest {
  /** ids are the leads to delete, at most 1000. */
  @minItems(1)
  @Onkt.encode("string")
  ids: int64[];
}

/** DeleteLeadsResponse reports how many leads were removed. */
model DeleteLeadsResponse {
  /** deleted is how many leads were removed; converted ones are kept. */
  deleted: int64;
}

/** ConvertLeadRequest qualifies a lead. */
model ConvertLeadRequest {
  /** id is the lead id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body says whether and how to open a deal. */
  @body
  body: ConvertLeadBody;
}

/**
 * ConvertLeadBody opens a deal for the qualified lead when create_deal is
 * set; its fields default from the lead.
 */
model ConvertLeadBody {
  /** create_deal opens a deal in the pipeline. */
  create_deal: boolean;

  /** deal holds the deal's fields. */
  deal?: DealFields;
}

/** ConvertLeadResponse returns the qualified lead and its deal. */
model ConvertLeadResponse {
  /** lead is the qualified lead. */
  lead: Lead;

  /** deal is the opened deal, when one was asked for. */
  deal?: Deal;
}

/** ListRepsRequest takes no parameters. */
model ListRepsRequest {}

/** Rep is a teammate work can be assigned to. */
model Rep {
  /** id is the user id. */
  @encode(string)
  id: int64;

  /** name is the teammate's display name. */
  name: string;
}

/** ListRepsResponse lists the business's assignable teammates. */
model ListRepsResponse {
  /** reps are the owner and active members. */
  reps: Rep[];
}

// ---------------------------------------------------------------------------
// Activities and tasks
// ---------------------------------------------------------------------------
/** Activity is one logged touch: a call, email, meeting, SMS or note. */
model Activity {
  /** id is the activity id. */
  @encode(string)
  id: int64;

  /** lead_id is the lead it was logged on. */
  @encode(string)
  lead_id?: int64;

  /** deal_id is the deal it was logged on. */
  @encode(string)
  deal_id?: int64;

  /** type is call, email, meeting, sms or note. */
  type: string;

  /**
   * outcome is a call's disposition: connected, no_answer, voicemail, busy,
   * wrong_number, callback, meeting_booked or not_interested.
   */
  outcome: string;

  /** notes are what was said or done. */
  notes: string;

  /** duration_seconds is a call's length. */
  duration_seconds?: int32;

  /** created_by is who logged it. */
  @encode(string)
  created_by: int64;

  /** occurred_at is when the touch happened. */
  @encode(string)
  occurred_at: int64;
}

/** SalesTaskFields create or edit a task. */
model SalesTaskFields {
  /** lead_id attaches the task to a lead (creation only). */
  @encode(string)
  lead_id?: int64;

  /** deal_id attaches the task to a deal (creation only). */
  @encode(string)
  deal_id?: int64;

  /** title says what to do. */
  title: string;

  /** type is call, email, meeting or todo. */
  type: string;

  /** due_date is when it is due (YYYY-MM-DD). */
  @Onkt.required
  due_date: string;

  /** assigned_to is who does it; the creator when absent. */
  @encode(string)
  assigned_to?: int64;
}

/** SalesTask is a follow-up due on a date. */
model SalesTask {
  /** id is the task id. */
  @encode(string)
  id: int64;

  /** lead_id is the lead it is about. */
  @encode(string)
  lead_id?: int64;

  /** deal_id is the deal it is about. */
  @encode(string)
  deal_id?: int64;

  /** title says what to do. */
  title: string;

  /** type is call, email, meeting or todo. */
  type: string;

  /** due_date is when it is due. */
  due_date: string;

  /** assigned_to is who does it. */
  @encode(string)
  assigned_to?: int64;

  /** completed_at is when it was done; absent while open. */
  @encode(string)
  completed_at?: int64;

  /** lead_name is the lead's person, for worklists. */
  lead_name: string;

  /** company_name is the lead's company. */
  company_name: string;

  /** phone is the lead's phone, so a call task can be dialled from the list. */
  phone: string;

  /** deal_title is the deal's title. */
  deal_title: string;

  /** created_at is when the task was created. */
  @encode(string)
  created_at: int64;
}

/** LogActivityRequest logs a touch on a lead or a deal. */
model LogActivityRequest {
  /** lead_id is the lead it was logged on. */
  @encode(string)
  lead_id?: int64;

  /** deal_id is the deal it was logged on. */
  @encode(string)
  deal_id?: int64;

  /** type is call, email, meeting, sms or note. */
  @Onkt.required
  type: string;

  /** outcome is required for calls. */
  outcome: string;

  /** notes are what was said or done. */
  notes: string;

  /** duration_seconds is a call's length. */
  duration_seconds?: int32;

  /** occurred_at is when it happened; now when absent. */
  @encode(string)
  occurred_at?: int64;

  /** follow_up schedules the next touch as a task. */
  follow_up?: SalesTaskFields;
}

/** LogActivityResponse returns the activity and any follow-up task. */
model LogActivityResponse {
  /** activity is the logged touch. */
  activity: Activity;

  /** follow_up is the scheduled task. */
  follow_up?: SalesTask;
}

/** ListLeadActivitiesRequest pages a lead's timeline, newest first. */
model ListLeadActivitiesRequest {
  /** lead_id is the lead. */
  @minValueExclusive(0)
  @encode(string)
  lead_id: int64;

  /** cursor continues from a previous page. */
  @query("cursor")
  cursor: string;

  /** limit is the page size. */
  @query("limit")
  limit: int32;
}

/**
 * ListDealActivitiesRequest pages a deal's timeline, including the work on
 * the lead it came from.
 */
model ListDealActivitiesRequest {
  /** deal_id is the deal. */
  @minValueExclusive(0)
  @encode(string)
  deal_id: int64;

  /** cursor continues from a previous page. */
  @query("cursor")
  cursor: string;

  /** limit is the page size. */
  @query("limit")
  limit: int32;
}

/** ListActivitiesResponse is one page of a timeline. */
model ListActivitiesResponse {
  /** activities is the page, newest first. */
  activities: Activity[];

  /** next_cursor continues the timeline; empty at its start. */
  next_cursor: string;
}

/** DeleteActivityRequest removes a logged activity. */
model DeleteActivityRequest {
  /** id is the activity id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** DeleteActivityResponse confirms the removal. */
model DeleteActivityResponse {
  /** deleted is true once the activity is gone. */
  deleted: boolean;
}

/** ListSalesTasksRequest pages tasks. */
model ListSalesTasksRequest {
  /** scope is open (earliest due first) or completed (latest first). */
  @query("scope")
  scope: string;

  /** assigned_to filters to one person's tasks; 0 selects unassigned. */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** due is overdue, today, due or upcoming, for open tasks. */
  @query("due")
  due: string;

  /** lead_id filters to one lead's tasks. */
  @query("lead_id")
  @encode(string)
  lead_id?: int64;

  /** deal_id filters to one deal's tasks. */
  @query("deal_id")
  @encode(string)
  deal_id?: int64;

  /** cursor continues from a previous page. */
  @query("cursor")
  cursor: string;

  /** limit is the page size. */
  @query("limit")
  limit: int32;

  /** today is the caller's date. */
  @query("today")
  today: string;
}

/** ListSalesTasksResponse is one page of tasks. */
model ListSalesTasksResponse {
  /** tasks is the page. */
  tasks: SalesTask[];

  /** next_cursor continues the list; empty on the last page. */
  next_cursor: string;
}

/** GetSalesTaskCountsRequest selects whose open work to count. */
model GetSalesTaskCountsRequest {
  /** assigned_to limits the counts to one person; everyone when absent. */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** today is the caller's date. */
  @query("today")
  today: string;
}

/** GetSalesTaskCountsResponse counts open tasks by when they are due. */
model GetSalesTaskCountsResponse {
  /** overdue counts tasks due before today. */
  overdue: int64;

  /** due_today counts tasks due today. */
  due_today: int64;

  /** upcoming counts tasks due after today. */
  upcoming: int64;
}

/** CreateSalesTaskRequest schedules a task. */
model CreateSalesTaskRequest {
  /** task holds the fields. */
  task: SalesTaskFields;
}

/** SalesTaskResponse returns a task. */
model SalesTaskResponse {
  /** task is the record as stored. */
  task: SalesTask;
}

/** UpdateSalesTaskRequest edits a task. */
model UpdateSalesTaskRequest {
  /** id is the task id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new values. */
  @body
  body: SalesTaskFields;
}

/** CompleteSalesTaskRequest marks a task done or open again. */
model CompleteSalesTaskRequest {
  /** id is the task id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body says which. */
  @body
  body: CompleteSalesTaskBody;
}

/** CompleteSalesTaskBody says whether the task is done. */
model CompleteSalesTaskBody {
  /** done is true to complete, false to reopen. */
  done: boolean;
}

/** DeleteSalesTaskRequest removes a task. */
model DeleteSalesTaskRequest {
  /** id is the task id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** DeleteSalesTaskResponse confirms the removal. */
model DeleteSalesTaskResponse {
  /** deleted is true once the task is gone. */
  deleted: boolean;
}

// ---------------------------------------------------------------------------
// Pipelines and deals
// ---------------------------------------------------------------------------
/** Stage is a step of a pipeline. */
model Stage {
  /** id is the stage id. */
  @encode(string)
  id: int64;

  /** pipeline_id is the pipeline it belongs to. */
  @encode(string)
  pipeline_id: int64;

  /** name is what the stage is called. */
  name: string;

  /** position orders the stages. */
  position: int32;

  /** probability (0-100) is the chance a deal in the stage is won. */
  probability: int32;

  /** kind is open, won or lost. */
  kind: string;
}

/** Pipeline is one sales motion's ordered stages. */
model Pipeline {
  /** id is the pipeline id. */
  @encode(string)
  id: int64;

  /** name is what the pipeline is called. */
  name: string;

  /** is_default marks the business's default pipeline. */
  is_default: boolean;

  /** stages are in order. */
  stages: Stage[];
}

/** StageFields describe a stage in a pipeline edit. */
model StageFields {
  /** id keeps an existing stage; absent adds one. */
  @encode(string)
  id?: int64;

  /** name is what the stage is called. */
  @Onkt.required
  name: string;

  /** probability (0-100); fixed at 100 for won and 0 for lost. */
  probability: int32;

  /** kind is open, won or lost; fixed once created. */
  @Onkt.required
  kind: string;
}

/** PipelineFields create or replace a pipeline. Stages left out are removed. */
model PipelineFields {
  /** name is what the pipeline is called. */
  @Onkt.required
  name: string;

  /**
   * stages are in order; a pipeline needs open stages and one won and one
   * lost stage.
   */
  @minItems(3)
  stages: StageFields[];
}

/** ListPipelinesRequest takes no parameters. */
model ListPipelinesRequest {}

/** ListPipelinesResponse lists the business's pipelines. */
model ListPipelinesResponse {
  /** pipelines are in order, with their stages. */
  pipelines: Pipeline[];
}

/** CreatePipelineRequest adds a pipeline. */
model CreatePipelineRequest {
  /** pipeline holds the fields. */
  pipeline: PipelineFields;
}

/** PipelineResponse returns a pipeline. */
model PipelineResponse {
  /** pipeline is the record as stored. */
  pipeline: Pipeline;
}

/** UpdatePipelineRequest renames a pipeline and replaces its stages. */
model UpdatePipelineRequest {
  /** id is the pipeline id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new values. */
  @body
  body: PipelineFields;
}

/** DeletePipelineRequest removes an empty pipeline. */
model DeletePipelineRequest {
  /** id is the pipeline id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** DeletePipelineResponse confirms the removal. */
model DeletePipelineResponse {
  /** deleted is true once the pipeline is gone. */
  deleted: boolean;
}

/** Deal is one opportunity in a pipeline. */
model Deal {
  /** id is the deal id. */
  @encode(string)
  id: int64;

  /** pipeline_id is the deal's pipeline. */
  @encode(string)
  pipeline_id: int64;

  /** stage_id is the deal's stage. */
  @encode(string)
  stage_id: int64;

  /** title names the deal. */
  title: string;

  /** lead_id is the lead it came from. */
  @encode(string)
  lead_id?: int64;

  /** customer_id is the CRM customer the deal is with. */
  @encode(string)
  customer_id?: int64;

  /** company_id is the company record the deal is with. */
  @encode(string)
  company_id?: int64;

  /** contact_name is the person the deal is with. */
  contact_name: string;

  /** company_name is the company the deal is with. */
  company_name: string;

  /** value is what would be charged per billing period (decimal string). */
  value: string;

  /** billing is one_time, monthly or yearly. */
  billing: string;

  /** amount is the first-year value pipeline totals use (decimal string). */
  amount: string;

  /** expected_close_date is when the deal should close. */
  expected_close_date?: string;

  /** assigned_to is the rep who owns the deal. */
  @encode(string)
  assigned_to?: int64;

  /** source is where the deal came from. */
  source: string;

  /** status is open, won or lost. */
  status: string;

  /** lost_reason says why a lost deal was lost. */
  lost_reason: string;

  /** closed_at is when the deal was won or lost. */
  @encode(string)
  closed_at?: int64;

  /** notes are free-text notes. */
  notes: string;

  /** last_activity_at is when the deal was last worked. */
  @encode(string)
  last_activity_at?: int64;

  /** next_follow_up is the earliest open task's due date. */
  next_follow_up?: string;

  /** stage_changed_at is when the deal entered its stage. */
  @encode(string)
  stage_changed_at: int64;

  /** created_at is when the deal was opened. */
  @encode(string)
  created_at: int64;
}

/** DealFields are a deal's editable fields. */
model DealFields {
  /** title names the deal; the company or contact when empty. */
  title: string;

  /** pipeline_id places a new deal; the default pipeline when absent. */
  @encode(string)
  pipeline_id?: int64;

  /** stage_id places a new deal; the pipeline's first open stage when absent. */
  @encode(string)
  stage_id?: int64;

  /** contact_name is the person the deal is with. */
  contact_name: string;

  /** company_name is the company the deal is with. */
  company_name: string;

  /** value is what would be charged per billing period (decimal string). */
  value: string;

  /** billing is one_time, monthly or yearly. */
  billing: string;

  /** expected_close_date is when the deal should close. */
  expected_close_date: string;

  /** assigned_to is the rep who owns the deal. */
  @encode(string)
  assigned_to?: int64;

  /** source is where the deal came from. */
  source: string;

  /** notes are free-text notes. */
  notes: string;
}

/** ListDealsRequest selects one page of deals for the table view. */
model ListDealsRequest {
  /** pipeline_id filters to one pipeline. */
  @query("pipeline_id")
  @encode(string)
  pipeline_id?: int64;

  /** stage_id filters to one stage. */
  @query("stage_id")
  @encode(string)
  stage_id?: int64;

  /** statuses is a comma-separated list of open, won and lost; all when empty. */
  @query("statuses")
  statuses: string;

  /** assigned_to filters to one rep's deals; 0 selects unassigned deals. */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** search matches the title, company and contact. */
  @query("search")
  search: string;

  /** sort is newest, close_date or value. */
  @query("sort")
  sort: string;

  /** cursor continues from a previous page. */
  @query("cursor")
  cursor: string;

  /** limit is the page size, at most 200. */
  @query("limit")
  limit: int32;
}

/** ListDealsResponse is one page of deals. */
model ListDealsResponse {
  /** deals is the page. */
  deals: Deal[];

  /** next_cursor continues the list; empty on the last page. */
  next_cursor: string;

  /** total is how many deals match the filters. */
  total: int64;
}

/** GetDealBoardRequest selects a pipeline's board. */
model GetDealBoardRequest {
  /** pipeline_id is the pipeline; the default when absent. */
  @query("pipeline_id")
  @encode(string)
  pipeline_id?: int64;

  /** assigned_to filters to one rep's deals; 0 selects unassigned deals. */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** search matches the title, company and contact. */
  @query("search")
  search: string;
}

/** BoardColumn is a stage's header on the board. */
model BoardColumn {
  /** stage is the column's stage. */
  stage: Stage;

  /** deals counts the stage's deals. */
  deals: int64;

  /** amount is their first-year value (decimal string). */
  amount: string;

  /** weighted_amount is amount times the stage's probability. */
  weighted_amount: string;
}

/**
 * GetDealBoardResponse is a pipeline with its column totals. Each column's
 * deals are paged with listStageDeals.
 */
model GetDealBoardResponse {
  /** pipeline is the board's pipeline. */
  pipeline: Pipeline;

  /** columns are the stages in order with their totals. */
  columns: BoardColumn[];
}

/** ListStageDealsRequest pages one board column, newest first. */
model ListStageDealsRequest {
  /** stage_id is the column's stage. */
  @minValueExclusive(0)
  @encode(string)
  stage_id: int64;

  /** assigned_to filters to one rep's deals; 0 selects unassigned deals. */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** search matches the title, company and contact. */
  @query("search")
  search: string;

  /** cursor continues from a previous page. */
  @query("cursor")
  cursor: string;

  /** limit is the page size. */
  @query("limit")
  limit: int32;
}

/** ListStageDealsResponse is one page of a column. */
model ListStageDealsResponse {
  /** deals is the page. */
  deals: Deal[];

  /** next_cursor continues the column; empty at its end. */
  next_cursor: string;
}

/** GetDealStatsRequest selects a pipeline's header stats. */
model GetDealStatsRequest {
  /** pipeline_id is the pipeline; the default when absent. */
  @query("pipeline_id")
  @encode(string)
  pipeline_id?: int64;

  /** assigned_to limits the stats to one rep. */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** today is the caller's date, for "this month". */
  @query("today")
  today: string;

  /** utc_offset_minutes is the caller's offset from UTC. */
  @query("utc_offset_minutes")
  utc_offset_minutes: int32;
}

/** GetDealStatsResponse summarises a pipeline. */
model GetSalesActivityRequest {
  /** days is the window, ending today: 7, 30 or 90 (30 otherwise). */
  @query("days")
  days: int32;

  /**
   * assigned_to limits the activity to one rep; 0 selects unassigned work.
   * Calls count for whoever logged them.
   */
  @query("assigned_to")
  @encode(string)
  assigned_to?: int64;

  /** today is the caller's date. */
  @query("today")
  today: string;

  /** utc_offset_minutes is the caller's offset from UTC. */
  @query("utc_offset_minutes")
  utc_offset_minutes: int32;
}

model SalesActivityDay {
  /** date is the caller's day (YYYY-MM-DD). */
  date: string;

  /** calls counts calls logged that day. */
  calls: int64;

  /** reached counts those calls that got through to the person. */
  reached: int64;

  /** leads_added counts leads created that day. */
  leads_added: int64;

  /** deals_won counts deals won that day. */
  deals_won: int64;
}

model RepActivity {
  /** rep_id is the user id; 0 is unassigned work. */
  @encode(string)
  rep_id: int64;

  /** calls counts the calls they logged. */
  calls: int64;

  /** reached counts those calls that got through. */
  reached: int64;

  /** leads_qualified counts their leads qualified in the window. */
  leads_qualified: int64;

  /** deals_won counts their deals won in the window. */
  deals_won: int64;

  /** deals_lost counts their deals lost in the window. */
  deals_lost: int64;

  /** won_amount is the first-year value of the deals won (decimal string). */
  won_amount: string;
}

model GetSalesActivityResponse {
  /** days covers every day of the window in order, quiet days included. */
  days: SalesActivityDay[];

  /** calls counts the window's calls. */
  calls: int64;

  /** reached counts those that got through. */
  reached: int64;

  /** leads_added counts leads created in the window. */
  leads_added: int64;

  /** leads_qualified counts leads qualified in the window. */
  leads_qualified: int64;

  /** deals_won counts deals won in the window. */
  deals_won: int64;

  /** deals_lost counts deals lost in the window. */
  deals_lost: int64;

  /** won_amount is the first-year value of the deals won (decimal string). */
  won_amount: string;

  /** reps is each person's share, busiest first. */
  reps: RepActivity[];
}

model GetDealStatsResponse {
  /** open_deals counts open deals. */
  open_deals: int64;

  /** open_amount is their first-year value (decimal string). */
  open_amount: string;

  /** weighted_amount is that value weighted by stage probability. */
  weighted_amount: string;

  /** won_this_month counts deals won this calendar month. */
  won_this_month: int64;

  /** won_amount_this_month is their first-year value. */
  won_amount_this_month: string;

  /**
   * win_rate is the share of deals closed in the last 90 days that were won
   * (0-100).
   */
  win_rate: float64;

  /** average_days_to_win is how long those won deals took. */
  average_days_to_win: float64;

  /** average_won_amount is their average first-year value. */
  average_won_amount: string;
}

/** GetDealRequest selects one deal. */
model GetDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** GetDealResponse returns a deal and its contract. */
model GetDealResponse {
  /** deal is the record. */
  deal: Deal;

  /** contract is the contract drafted when the deal was won. */
  contract?: Contract;
}

/** CreateDealRequest opens a deal without a lead. */
model CreateDealRequest {
  /** deal holds the fields. */
  deal: DealFields;
}

/** DealResponse returns a deal. */
model DealResponse {
  /** deal is the record as stored. */
  deal: Deal;
}

/** UpdateDealRequest edits a deal's fields. */
model UpdateDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new values. */
  @body
  body: DealFields;
}

/** MoveDealRequest puts a deal in another open stage. */
model MoveDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body names the stage. */
  @body
  body: MoveDealBody;
}

/** MoveDealBody names the target stage. */
model MoveDealBody {
  /** stage_id is an open stage, in the deal's pipeline or another. */
  @minValueExclusive(0)
  @encode(string)
  stage_id: int64;
}

/** WinDealRequest closes a deal as won. */
model WinDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body optionally drafts the contract. */
  @body
  body: WinDealBody;
}

/**
 * WinDealBody drafts the won deal's contract when contract is present; its
 * fields default from the deal.
 */
model WinDealBody {
  /** contract holds the contract's terms. */
  contract?: ContractFields;
}

/** ContractFields are a drafted contract's terms. */
model ContractFields {
  /** plan is what was sold; the deal title when empty. */
  plan: string;

  /** value is charged per billing period; the deal's value when empty. */
  value: string;

  /** billing is one_time, monthly or yearly; the deal's when empty. */
  billing: string;

  /** seats is how many users the contract covers. */
  seats: int32;

  /** start_date is when the term starts; today when absent. */
  start_date?: string;

  /** end_date is when the term ends. */
  end_date?: string;

  /** signed_at records the signature now (YYYY-MM-DD). */
  signed_at?: string;

  /** signed_by_name names the customer-side signer. */
  signed_by_name: string;

  /** notes are free-text notes. */
  notes: string;
}

/** WinDealResponse returns the won deal and its contract. */
model WinDealResponse {
  /** deal is the won deal. */
  deal: Deal;

  /** contract is the drafted contract. */
  contract?: Contract;
}

/** LoseDealRequest closes a deal as lost. */
model LoseDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body gives the reason. */
  @body
  body: LoseDealBody;
}

/** LoseDealBody gives the reason. */
model LoseDealBody {
  /** reason says why the deal was lost. */
  @Onkt.required
  reason: string;
}

/** ReopenDealRequest returns a closed deal to an open stage. */
model ReopenDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body names the stage. */
  @body
  body: ReopenDealBody;
}

/** ReopenDealBody names the stage to reopen into. */
model ReopenDealBody {
  /** stage_id is an open stage; the pipeline's first when absent. */
  @encode(string)
  stage_id?: int64;
}

/** DeleteDealRequest removes a deal without a contract. */
model DeleteDealRequest {
  /** id is the deal id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** DeleteDealResponse confirms the removal. */
model DeleteDealResponse {
  /** deleted is true once the deal is gone. */
  deleted: boolean;
}

// ---------------------------------------------------------------------------
// Lead imports
// ---------------------------------------------------------------------------
/** LeadImport is one uploaded file's progress and outcome. */
model LeadImport {
  /** id is the import id. */
  @encode(string)
  id: int64;

  /** file_name is the uploaded file's name. */
  file_name: string;

  /** total_rows is how many rows the file has. */
  total_rows: int32;

  /** created counts rows that became new leads. */
  created: int32;

  /** updated counts rows that filled in an existing lead. */
  updated: int32;

  /** skipped counts rows that duplicated an existing lead. */
  skipped: int32;

  /** failed counts rows that could not be imported. */
  failed: int32;

  /** status is running, completed or undone. */
  status: string;

  /** created_by is who started the import. */
  @encode(string)
  created_by: int64;

  /** created_at is when it started. */
  @encode(string)
  created_at: int64;

  /** completed_at is when it finished or was undone. */
  @encode(string)
  completed_at?: int64;
}

/** StartLeadImportRequest opens an import before its rows are sent. */
model StartLeadImportRequest {
  /** file_name is the uploaded file's name. */
  file_name: string;

  /** total_rows is how many rows will be sent, at most 200,000. */
  @minValueExclusive(0)
  total_rows: int32;
}

/** LeadImportResponse returns an import. */
model LeadImportResponse {
  /** import is the record. */
  `import`: LeadImport;
}

/** ImportLeadRowsRequest sends one chunk of an import's rows. */
model ImportLeadRowsRequest {
  /** id is the import id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the rows. */
  @body
  body: ImportLeadRowsBody;
}

/** ImportLeadRow is one spreadsheet row. */
model ImportLeadRow {
  /** row is the row's line in the file, echoed in results. */
  row: int32;

  /** lead holds the mapped fields. */
  lead: LeadFields;
}

/** ImportLeadRowsBody is a chunk of at most 500 rows and how to handle them. */
model ImportLeadRowsBody {
  /** rows are the chunk. */
  @minItems(1)
  rows: ImportLeadRow[];

  /**
   * duplicates is skip (default) or fill: fill only completes the missing
   * fields of the existing lead.
   */
  duplicates: string;

  /** assigned_to gives rows without a rep to this one. */
  @encode(string)
  assigned_to?: int64;

  /** source labels rows without a source. */
  source: string;
}

/** ImportLeadRowResult reports one row back. */
model ImportLeadRowResult {
  /** row is the row's line in the file. */
  row: int32;

  /** outcome is created, updated, skipped or failed. */
  outcome: string;

  /** lead_id is the created or matched lead. */
  @encode(string)
  lead_id?: int64;

  /** error says why a failed row was rejected. */
  error: string;
}

/** ImportLeadRowsResponse reports the chunk's rows. */
model ImportLeadRowsResponse {
  /** results are one per row. */
  results: ImportLeadRowResult[];
}

/** FinishLeadImportRequest marks an import complete. */
model FinishLeadImportRequest {
  /** id is the import id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** ListLeadImportsRequest takes no parameters. */
model ListLeadImportsRequest {}

/** ListLeadImportsResponse lists recent imports. */
model ListLeadImportsResponse {
  /** imports are the 20 most recent. */
  imports: LeadImport[];
}

/** UndoLeadImportRequest removes an import's untouched leads. */
model UndoLeadImportRequest {
  /** id is the import id. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

/** UndoLeadImportResponse says what undoing removed and kept. */
model UndoLeadImportResponse {
  /** removed counts leads deleted. */
  removed: int64;

  /** kept counts leads kept because someone has worked them since. */
  kept: int64;
}
