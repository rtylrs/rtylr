import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.sales.v1;

/**
 * LeadService covers prospecting: the lead list and its stats, editing,
 * bulk changes and conversion into a deal.
 */
@route("/crm/sales/v1")
interface LeadService {
  /** Returns one page of leads. */
  @get
  @route("/leads")
  listLeads(...ListLeadsRequest): ListLeadsResponse;

  /**
   * Summarises prospecting for the leads page header, for the same filters
   * as the list (statuses aside).
   */
  @get
  @route("/leads/stats")
  getLeadStats(...GetLeadStatsRequest): GetLeadStatsResponse;

  /** Lists the distinct sources, industries, sizes and places leads have. */
  @get
  @route("/leads/facets")
  listLeadFacets(...ListLeadFacetsRequest): ListLeadFacetsResponse;

  /** Returns a single lead. */
  @get
  @route("/leads/{id}")
  getLead(...GetLeadRequest): GetLeadResponse;

  /** Adds a lead. */
  @post
  @route("/leads")
  createLead(...CreateLeadRequest): CreateLeadResponse;

  /** Replaces a lead's fields and status. */
  @put
  @route("/leads/{id}")
  updateLead(...UpdateLeadRequest): UpdateLeadResponse;

  /** Removes an unconverted lead with its tasks and activities. */
  @delete
  @route("/leads/{id}")
  deleteLead(...DeleteLeadRequest): DeleteLeadResponse;

  /** Assigns or re-statuses many leads. */
  @put
  @route("/leads")
  updateLeads(...UpdateLeadsRequest): UpdateLeadsResponse;

  /** Removes many unconverted leads. */
  @post
  @route("/leads/delete")
  deleteLeads(...DeleteLeadsRequest): DeleteLeadsResponse;

  /** Qualifies a lead into a person, a company and optionally a deal. */
  @put
  @route("/leads/{id}/conversion")
  convertLead(...ConvertLeadRequest): ConvertLeadResponse;

  /** Lists the teammates work can be assigned to. */
  @get
  @route("/reps")
  listReps(...ListRepsRequest): ListRepsResponse;
}

/**
 * ActivityService logs calls, emails, meetings, SMS and notes, and schedules
 * the follow-ups that drive each rep's worklist.
 */
@route("/crm/sales/v1")
interface ActivityService {
  /** Logs a touch, optionally scheduling the next one. */
  @post
  @route("/activities")
  logActivity(...LogActivityRequest): LogActivityResponse;

  /** Pages a lead's timeline. */
  @get
  @route("/leads/{lead_id}/activities")
  listLeadActivities(...ListLeadActivitiesRequest): ListActivitiesResponse;

  /** Pages a deal's timeline. */
  @get
  @route("/deals/{deal_id}/activities")
  listDealActivities(...ListDealActivitiesRequest): ListActivitiesResponse;

  /** Removes a logged activity. */
  @delete
  @route("/activities/{id}")
  deleteActivity(...DeleteActivityRequest): DeleteActivityResponse;

  /** Pages tasks. */
  @get
  @route("/tasks")
  listTasks(...ListSalesTasksRequest): ListSalesTasksResponse;

  /** Counts open tasks by when they are due. */
  @get
  @route("/tasks/counts")
  getTaskCounts(...GetSalesTaskCountsRequest): GetSalesTaskCountsResponse;

  /** Schedules a task. */
  @post
  @route("/tasks")
  createTask(...CreateSalesTaskRequest): SalesTaskResponse;

  /** Edits a task. */
  @put
  @route("/tasks/{id}")
  updateTask(...UpdateSalesTaskRequest): SalesTaskResponse;

  /** Marks a task done or open again. */
  @put
  @route("/tasks/{id}/completion")
  completeTask(...CompleteSalesTaskRequest): SalesTaskResponse;

  /** Removes a task. */
  @delete
  @route("/tasks/{id}")
  deleteTask(...DeleteSalesTaskRequest): DeleteSalesTaskResponse;
}

/** DealService covers the pipeline: boards, deal tables, stats and closing. */
@route("/crm/sales/v1")
interface DealService {
  /** Lists the business's pipelines, creating the default on first use. */
  @get
  @route("/pipelines")
  listPipelines(...ListPipelinesRequest): ListPipelinesResponse;

  /** Adds a pipeline. */
  @post
  @route("/pipelines")
  createPipeline(...CreatePipelineRequest): PipelineResponse;

  /** Renames a pipeline and replaces its stages. */
  @put
  @route("/pipelines/{id}")
  updatePipeline(...UpdatePipelineRequest): PipelineResponse;

  /** Removes an empty pipeline. */
  @delete
  @route("/pipelines/{id}")
  deletePipeline(...DeletePipelineRequest): DeletePipelineResponse;

  /** Returns a pipeline's board header: each stage's totals. */
  @get
  @route("/deals/board")
  getDealBoard(...GetDealBoardRequest): GetDealBoardResponse;

  /** Pages one board column. */
  @get
  @route("/stages/{stage_id}/deals")
  listStageDeals(...ListStageDealsRequest): ListStageDealsResponse;

  /** Summarises a pipeline. */
  @get
  @route("/deals/stats")
  getDealStats(...GetDealStatsRequest): GetDealStatsResponse;

  /**
   * Returns calls, new leads and won deals per day over a window, with its
   * totals and each rep's share, for the dashboard.
   */
  @get
  @route("/activity")
  getSalesActivity(...GetSalesActivityRequest): GetSalesActivityResponse;

  /** Returns one page of deals. */
  @get
  @route("/deals")
  listDeals(...ListDealsRequest): ListDealsResponse;

  /** Returns a single deal with its contract. */
  @get
  @route("/deals/{id}")
  getDeal(...GetDealRequest): GetDealResponse;

  /** Opens a deal without a lead. */
  @post
  @route("/deals")
  createDeal(...CreateDealRequest): DealResponse;

  /** Edits a deal's fields. */
  @put
  @route("/deals/{id}")
  updateDeal(...UpdateDealRequest): DealResponse;

  /** Moves a deal to another open stage. */
  @put
  @route("/deals/{id}/stage")
  moveDeal(...MoveDealRequest): DealResponse;

  /** Closes a deal as won, optionally drafting its contract. */
  @put
  @route("/deals/{id}/win")
  winDeal(...WinDealRequest): WinDealResponse;

  /** Closes a deal as lost. */
  @put
  @route("/deals/{id}/lose")
  loseDeal(...LoseDealRequest): DealResponse;

  /** Reopens a closed deal. */
  @put
  @route("/deals/{id}/reopen")
  reopenDeal(...ReopenDealRequest): DealResponse;

  /** Removes a deal without a contract. */
  @delete
  @route("/deals/{id}")
  deleteDeal(...DeleteDealRequest): DeleteDealResponse;
}

/** LeadImportService loads spreadsheets of leads in chunks. */
@route("/crm/sales/v1")
interface LeadImportService {
  /** Opens an import. */
  @post
  @route("/lead-imports")
  startLeadImport(...StartLeadImportRequest): LeadImportResponse;

  /** Imports one chunk of rows. */
  @post
  @route("/lead-imports/{id}/rows")
  importLeadRows(...ImportLeadRowsRequest): ImportLeadRowsResponse;

  /** Marks an import complete. */
  @put
  @route("/lead-imports/{id}")
  finishLeadImport(...FinishLeadImportRequest): LeadImportResponse;

  /** Lists recent imports. */
  @get
  @route("/lead-imports")
  listLeadImports(...ListLeadImportsRequest): ListLeadImportsResponse;

  /** Removes an import's untouched leads. */
  @delete
  @route("/lead-imports/{id}")
  undoLeadImport(...UndoLeadImportRequest): UndoLeadImportResponse;
}

/**
 * ContractService covers contracts: the read surface other apps need
 * (Finance prefills invoices from signed contracts) plus drafting and
 * amending.
 */
@route("/crm/sales/v1")
interface ContractService {
  /** Returns one page of contracts for this business. */
  @get
  @route("/contracts")
  listContracts(...ListContractsRequest): ListContractsResponse;

  /** Returns a single contract by id. */
  @get
  @route("/contracts/{id}")
  getContract(...GetContractRequest): GetContractResponse;

  /** Drafts a contract, usually for a won deal. */
  @post
  @route("/contracts")
  createContract(...CreateContractRequest): CreateContractResponse;

  /** Replaces a contract's terms. */
  @put
  @route("/contracts/{id}")
  updateContract(...UpdateContractRequest): UpdateContractResponse;
}
