import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.modifier.v1;

/** ModifierService defines POS menu item modifier endpoints. */
@route("/pos/modifier/v1")
interface ModifierService {
  /** Returns modifiers for the active business. */
  @get
  @route("/modifiers")
  listModifiers(...ListModifiersRequest): ListModifiersResponse;

  /** Creates a new modifier. */
  @post
  @route("/modifiers")
  createModifier(...CreateModifierRequest): CreateModifierResponse;

  /** Updates an existing modifier. */
  @put
  @route("/modifiers/{id}")
  updateModifier(...UpdateModifierRequest): UpdateModifierResponse;

  /** Removes a modifier by ID. */
  @delete
  @route("/modifiers/{id}")
  deleteModifier(...DeleteModifierRequest): DeleteModifierResponse;
}
