import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace website.lead.v1;

/**
 * Public lead capture for the marketing website. The routes keep the paths
 * the website already posts to.
 */
@route("/")
interface WebsiteLeadService {
  @post
  @route("/quote")
  @Onkt.meta("rate_limit", "10/m")
  submitQuote(...SubmitQuoteRequest): SubmitQuoteResponse;

  @post
  @route("/contact")
  @Onkt.meta("rate_limit", "10/m")
  contact(...ContactRequest): ContactResponse;

  @post
  @route("/newsletter")
  @Onkt.meta("rate_limit", "10/m")
  subscribeNewsletter(...NewsletterRequest): NewsletterResponse;

  @post
  @route("/analytics/event")
  @Onkt.meta("rate_limit", "60/m")
  recordAnalyticsEvent(...AnalyticsEventRequest): AnalyticsEventResponse;
}
