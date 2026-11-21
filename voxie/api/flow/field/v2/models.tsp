import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.field.v2;

/** FieldOption is one choice of a `select` or `multi_select` field. */
model FieldOption {
  /**
   * id identifies the option inside its field. Task values store this id,
   * never the label, so renaming an option keeps every value. The server
   * assigns it: send it empty for a new option and unchanged for an existing
   * one.
   */
  id: string;

  /** label is what the user sees. Unique within the field, ignoring case. */
  @Onkt.required
  label: string;

  /** color is a hex string. Optional, server default applies. */
  color: string;
}

/**
 * Field is a custom field defined on one project. Every task in the project
 * can carry one value for it (see `FlowTask.field_values` in flow.task.v2).
 */
model Field {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** project_id is the project this field belongs to. */
  @encode(string)
  project_id: int64;

  /** name is unique within the project, ignoring case. */
  name: string;

  /**
   * type is one of "text", "number", "date", "select", "multi_select",
   * "person", "url" or "checkbox". It cannot change after creation.
   */
  type: string;

  /**
   * options is the choice list of a select or multi_select field, in display
   * order. Empty for every other type.
   */
  options: FieldOption[];

  /**
   * required makes a value mandatory when a task is created in the project
   * and refuses clearing it later. It is not applied retroactively: a task
   * that predates the rule can still be edited without a value.
   */
  required: boolean;

  /**
   * default_value is the value a new task receives when it sends none, for
   * the single-value types. Same encoding as `TaskFieldValue.value`. Empty
   * means no default.
   */
  default_value: string;

  /** default_values is the default of a multi_select or person field. */
  default_values: string[];

  /** sort_order is the position within the project's fields, ascending. */
  sort_order: int32;

  /**
   * value_count is how many tasks hold a value for this field. Filled by
   * listFields so a delete can say what it will clear; zero on the other
   * responses.
   */
  value_count: int32;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model ListFieldsRequest {
  /** project_id scopes the listing to one project's fields. */
  @query("project_id")
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;
}

model ListFieldsResponse {
  /** fields is the project's fields in sort order. Empty when it has none. */
  fields: Field[];
}

model CreateFieldRequest {
  /**
   * project_id is the project to add the field to. A project holds at most
   * 50 fields.
   */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** name must be unique within the project. A duplicate is rejected with 409. */
  @Onkt.required
  name: string;

  /** type is one of the eight field types. */
  @Onkt.required
  type: string;

  /**
   * options is required for select and multi_select (1 to 100 entries) and
   * must be empty for every other type. Send each option without an id.
   */
  options: FieldOption[];

  /** required makes a value mandatory on task creation. */
  required: boolean;

  /**
   * default_value is the default for single-value types. For a select it is
   * the label of one of `options`, because option ids do not exist yet.
   */
  default_value: string;

  /**
   * default_values is the default for multi_select (option labels) or person
   * (user ids).
   */
  default_values: string[];
}

model CreateFieldResponse {
  /** field is the created field, with its option ids assigned. */
  field: Field;
}

model UpdateFieldRequest {
  /** id is the field identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  @body
  body: UpdateFieldBody;
}

/**
 * UpdateFieldBody is a replacement, so send the whole field. The type is not
 * part of it because it cannot change, and the position is changed through
 * reorderFields.
 */
model UpdateFieldBody {
  /** name must remain unique within the project. Renaming keeps every value. */
  @Onkt.required
  name: string;

  /**
   * options is the full new choice list in display order. An option sent
   * with its id is kept (rename or recolour); one sent without an id is new;
   * an existing option left out is removed, and it is cleared from every
   * task value in the same transaction.
   */
  options: FieldOption[];

  /** required makes a value mandatory on task creation. */
  required: boolean;

  /**
   * default_value is the default for single-value types. For a select it is
   * an option id from this same request, or the label of a new option.
   */
  default_value: string;

  /** default_values is the default for multi_select or person. */
  default_values: string[];
}

model UpdateFieldResponse {
  /** field is the updated field. */
  field: Field;
}

model ReorderFieldsRequest {
  /** project_id is the project whose fields are reordered. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /**
   * field_ids is every field in the project, in the new order. A missing,
   * extra, repeated or foreign id is rejected with 400.
   */
  @Onkt.required
  @Onkt.encode("string")
  field_ids: int64[];
}

model ReorderFieldsResponse {
  /** fields is the project's fields in their new order. */
  fields: Field[];
}

model DeleteFieldRequest {
  /** id is the field identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteFieldResponse {
  /** message is a human-readable result message. */
  message: string;

  /**
   * cleared_task_count is how many tasks had a value for the field. Those
   * values are deleted with it; the tasks themselves are untouched.
   */
  cleared_task_count: int32;
}
