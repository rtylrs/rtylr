import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.combo.v1;

/** ComboService defines POS combo endpoints. */
@route("/pos/combo/v1")
interface ComboService {
  /** Returns all combos for the active business. */
  @get
  @route("/combos")
  listCombos(...ListCombosRequest): ListCombosResponse;

  /** Creates a new combo. */
  @post
  @route("/combos")
  createCombo(...CreateComboRequest): CreateComboResponse;

  /** Updates an existing combo. */
  @put
  @route("/combos/{id}")
  updateCombo(...UpdateComboRequest): UpdateComboResponse;

  /** Removes a combo by ID. */
  @delete
  @route("/combos/{id}")
  deleteCombo(...DeleteComboRequest): DeleteComboResponse;
}
