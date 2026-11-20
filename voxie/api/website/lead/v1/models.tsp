import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace website.lead.v1;

/**
 * A quote request from the public website. Email or phone_number is
 * required, at least one, so a visitor can ask for a quote with only a phone.
 */
model SubmitQuoteRequest {
  @Onkt.required
  @minLength(1)
  @maxLength(200)
  name: string;

  @minLength(0)
  @maxLength(255)
  email: string;

  @minLength(0)
  @maxLength(50)
  phone_number: string;

  @minLength(0)
  @maxLength(200)
  company: string;

  @Onkt.required
  @minLength(1)
  @maxLength(200)
  service: string;

  @minLength(0)
  @maxLength(5000)
  project: string;

  @minLength(0)
  @maxLength(200)
  budget: string;

  @minLength(0)
  @maxLength(200)
  timeline: string;

  @minLength(0)
  @maxLength(200)
  source: string;
}

model SubmitQuoteResponse {
  @encode(string)
  id: int64;

  message: string;
}

/** A contact message from the public website. */
model ContactRequest {
  @Onkt.required
  @minLength(1)
  @maxLength(200)
  name: string;

  @Onkt.required
  @minLength(3)
  @maxLength(255)
  email: string;

  @minLength(0)
  @maxLength(50)
  phone_number: string;

  @minLength(0)
  @maxLength(5000)
  message: string;
}

model ContactResponse {
  message: string;
}

/** A newsletter signup from the public website or blog. */
model NewsletterRequest {
  @Onkt.required
  @minLength(3)
  @maxLength(255)
  email: string;

  @minLength(0)
  @maxLength(100)
  source: string;

  @minLength(0)
  @maxLength(255)
  post_slug: string;

  @minLength(0)
  @maxLength(255)
  post_title: string;

  @minLength(0)
  @maxLength(100)
  post_category: string;

  @minLength(0)
  @maxLength(10)
  language: string;
}

model NewsletterResponse {
  @encode(string)
  id: int64;

  message: string;
}

/** A best-effort analytics event; every field is optional. */
model AnalyticsEventRequest {
  @minLength(0)
  @maxLength(100)
  event: string;

  @minLength(0)
  @maxLength(500)
  page: string;

  /** Job ids arrive as numbers or strings depending on the page. */
  job_id: unknown;

  @minLength(0)
  @maxLength(500)
  job_title: string;
}

model AnalyticsEventResponse {
  logged: boolean;
}
