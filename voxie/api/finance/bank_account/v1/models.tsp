import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace finance.bank_account.v1;

/** BankAccount is a business's linked bank account. */
model BankAccount {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** bank_name is the bank name value for this contract. */
  bank_name: string;

  /**
   * account_type is the account type value for this contract.
   * checking, savings
   */
  account_type: string;

  /** last_four is the last four value for this contract. */
  last_four: string;

  /** balance is the balance value for this contract. */
  balance: float64;

  /** currency is the currency value for this contract. */
  currency: string;

  /** is_primary is the is primary value for this contract. */
  is_primary: boolean;

  /** created_at is when the record was created. */
  created_at: string;
}

model CreateBankAccountRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** bank_name is the bank name value for this contract. */
  @Onkt.required
  bank_name: string;

  /** account_type is the account type value for this contract. */
  @Onkt.required
  account_type: string;

  /** last_four is the last four value for this contract. */
  @Onkt.required
  last_four: string;

  /** balance is the balance value for this contract. */
  @Onkt.required
  @minValue(0)
  balance?: float64;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** is_primary is the is primary value for this contract. */
  is_primary: boolean;
}

model CreateBankAccountResponse {
  /** account is the account value for this contract. */
  account: BankAccount;
}

model GetBankAccountRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetBankAccountResponse {
  /** account is the account value for this contract. */
  account: BankAccount;
}

model ListBankAccountsRequest {
  /** page is the current result page. */
  @query("page")
  page: int32;

  /** limit is the maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** sort_by is the field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** sort_dir is the sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListBankAccountsResponse {
  /** accounts is the accounts value for this contract. */
  accounts: BankAccount[];

  /** pagination is the pagination value for this contract. */
  pagination: PaginationResponse;
}
