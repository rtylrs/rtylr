import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.account.v1;

/** AccountService defines chart-of-accounts endpoints. */
@route("/finance/account/v1")
interface AccountService {
  @get
  @route("/accounts")
  listAccounts(...ListAccountsRequest): ListAccountsResponse;

  @post
  @route("/accounts")
  createAccount(...CreateAccountRequest): CreateAccountResponse;

  @put
  @route("/accounts/{id}")
  updateAccount(...UpdateAccountRequest): UpdateAccountResponse;

  @delete
  @route("/accounts/{id}")
  deleteAccount(...DeleteAccountRequest): DeleteAccountResponse;

  @get
  @route("/chart-of-accounts")
  getChartOfAccounts(...GetChartOfAccountsRequest): GetChartOfAccountsResponse;
}
