import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.quote.v1;

/** QuoteService defines sales quote endpoints. */
@route("/finance/quote/v1")
interface QuoteService {
  @get
  @route("/quotes")
  listQuotes(...ListQuotesRequest): ListQuotesResponse;

  @post
  @route("/quotes")
  createQuote(...CreateQuoteRequest): CreateQuoteResponse;

  @get
  @route("/quotes/{id}")
  getQuote(...GetQuoteRequest): GetQuoteResponse;

  @patch
  @route("/quotes/{id}/status")
  updateQuoteStatus(...UpdateQuoteStatusRequest): UpdateQuoteStatusResponse;

  @post
  @route("/quotes/{id}/convert")
  convertQuote(...ConvertQuoteRequest): ConvertQuoteResponse;
}
