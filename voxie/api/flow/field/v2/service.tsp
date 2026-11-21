import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.field.v2;

/**
 * FieldService manages a project's custom field definitions.
 *
 * Task values are written through flow.task.v2 (`setTaskFieldValues` and the
 * `field_values` of `createTask`), because writing a value is editing a task.
 */
@route("/flow/field/v2")
interface FieldService {
  /** Returns one project's fields in sort order. */
  @get
  @route("/fields")
  listFields(...ListFieldsRequest): ListFieldsResponse;

  /** Creates a field at the end of a project's list. */
  @post
  @route("/fields")
  createField(...CreateFieldRequest): CreateFieldResponse;

  /** Sets the order of a project's whole field list in one write. */
  @put
  @route("/fields/order")
  reorderFields(...ReorderFieldsRequest): ReorderFieldsResponse;

  /** Replaces a field's name, options, required flag and default. */
  @put
  @route("/fields/{id}")
  updateField(...UpdateFieldRequest): UpdateFieldResponse;

  /** Deletes a field and every task value stored for it. */
  @delete
  @route("/fields/{id}")
  deleteField(...DeleteFieldRequest): DeleteFieldResponse;
}
