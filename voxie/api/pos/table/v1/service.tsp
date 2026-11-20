import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.table.v1;

/** TableService defines POS floor table endpoints. */
@route("/pos/table/v1")
interface TableService {
  /** Returns tables for the active business. */
  @get
  @route("/tables")
  listTables(...ListTablesRequest): ListTablesResponse;

  /** Creates a new table. */
  @post
  @route("/tables")
  createTable(...CreateTableRequest): CreateTableResponse;

  /** Updates an existing table. */
  @put
  @route("/tables/{id}")
  updateTable(...UpdateTableRequest): UpdateTableResponse;

  /** Removes a table by ID. */
  @delete
  @route("/tables/{id}")
  deleteTable(...DeleteTableRequest): DeleteTableResponse;
}
