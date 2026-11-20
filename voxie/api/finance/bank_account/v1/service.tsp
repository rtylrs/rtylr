import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.bank_account.v1;

/** BankAccountService defines bank account endpoints. */
@route("/finance/bank-account/v1")
interface BankAccountService {
  @get
  @route("/bank-accounts")
  listBankAccounts(...ListBankAccountsRequest): ListBankAccountsResponse;

  @post
  @route("/bank-accounts")
  createBankAccount(...CreateBankAccountRequest): CreateBankAccountResponse;

  @get
  @route("/bank-accounts/{id}")
  getBankAccount(...GetBankAccountRequest): GetBankAccountResponse;
}
