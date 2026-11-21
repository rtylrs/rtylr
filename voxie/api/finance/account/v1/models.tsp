import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.account.v1;

/** Account is a chart-of-accounts ledger account. */
model Account {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** type is the type value for this contract. */
  type: string;

  /** currency is the currency value for this contract. */
  currency: string;

  /** balance is the balance value for this contract. */
  balance: float64;

  /** is_active is the is active value for this contract. */
  is_active: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** code is the optional account code/number. */
  code: string;

  /**
   * current_balance is balance (the opening balance) plus every posted
   * journal line on the account, signed by the account type.
   */
  current_balance: float64;
}

/** ChartOfAccountsNode is a node in the hierarchical chart of accounts. */
model ChartOfAccountsNode {
  @encode(string)
  id: int64;

  name: string;
  code: string;
  type: string;
  children: ChartOfAccountsNode[];
}

model CreateAccountRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** type is the type value for this contract. */
  @Onkt.required
  type: string;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** balance is the balance value for this contract. */
  @Onkt.required
  @minValue(0)
  balance?: float64;

  /** is_active is the is active value for this contract. */
  is_active: boolean;

  /** code is the optional account code/number. */
  code: string;
}

model CreateAccountResponse {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model DeleteAccountRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteAccountResponse {
  message: string;
}

model GetChartOfAccountsRequest {}

model GetChartOfAccountsResponse {
  tree: ChartOfAccountsNode[];
}

model ListAccountsRequest {
  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

model ListAccountsResponse {
  /** accounts is the accounts value for this contract. */
  accounts: Account[];

  page: int64;
  page_size: int64;
  has_more: boolean;
}

model UpdateAccountBody {
  @Onkt.required
  name: string;

  @Onkt.required
  type: string;

  currency: string;

  @minValue(0)
  balance: float64;

  is_active: boolean;

  /** code is the optional account code/number. */
  code: string;
}

model UpdateAccountRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateAccountBody;
}

model UpdateAccountResponse {
  message: string;
}
