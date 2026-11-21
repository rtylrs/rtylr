import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace recruit.compat.v1;

/** Typed compatibility surface for the deployed Recruit routes. */
@route("/recruit")
interface CompatibilityService {
  @get
  @route("/careers")
  listCareers(...EmptyRequest): JsonResponse;

  @post
  @route("/careers")
  createCareer(...BodyRequest): JsonResponse;

  @put
  @route("/careers")
  updateCareer(...BodyRequest): JsonResponse;

  @delete
  @route("/careers")
  deleteCareer(...IdQueryRequest): JsonResponse;

  @get
  @route("/customization")
  getCustomization(...EmptyRequest): JsonResponse;

  @put
  @route("/customization")
  updateCustomization(...BodyRequest): JsonResponse;

  @get
  @route("/submissions")
  listSubmissions(...QueryRequest): JsonResponse;

  @put
  @route("/submissions/status")
  updateSubmissionStatus(...BodyRequest): JsonResponse;

  @get
  @route("/submissions/resume-url")
  getResumeURL(...QueryRequest): JsonResponse;

  @get
  @route("/sections")
  listSections(...EmptyRequest): JsonResponse;

  @post
  @route("/sections")
  createSection(...BodyRequest): JsonResponse;

  @put
  @route("/sections")
  updateSection(...BodyRequest): JsonResponse;

  @delete
  @route("/sections")
  deleteSection(...IdQueryRequest): JsonResponse;

  @put
  @route("/sections/reorder")
  reorderSections(...BodyRequest): JsonResponse;

  @get
  @route("/form-fields")
  listFormFields(...QueryRequest): JsonResponse;

  @post
  @route("/form-fields")
  createFormField(...BodyRequest): JsonResponse;

  @put
  @route("/form-fields")
  updateFormField(...BodyRequest): JsonResponse;

  @delete
  @route("/form-fields")
  deleteFormField(...IdQueryRequest): JsonResponse;

  @get
  @route("/public/careers")
  listPublicCareers(...QueryRequest): JsonResponse;

  @get
  @route("/public/career")
  getPublicCareer(...QueryRequest): JsonResponse;

  @get
  @route("/public/branding")
  getPublicBranding(...EmptyRequest): JsonResponse;

  @get
  @route("/submissions/export")
  exportSubmissions(...EmptyRequest): JsonResponse;

  // The public apply and resume upload routes run inside per-route rate
  // limits a contract method cannot apply, so their original handlers serve
  // them.
  @get
  @route("/public/og")
  @Onkt.http
  listPublicOg(...EmptyRequest): EmptyResponse;

  @post
  @route("/public/apply")
  @Onkt.http
  createPublicApply(...EmptyRequest): EmptyResponse;

  @post
  @route("/public/resume-upload-url")
  @Onkt.http
  createPublicResumeUploadUrl(...EmptyRequest): EmptyResponse;
}

/**
 * Voxhire recruiting routes. The admin routes run behind the Voxhire admin
 * session guard and the login, contact and apply routes behind per-route
 * rate limits, so they are \@http methods served by the original handlers.
 */
@route("/voxhire")
interface VoxhireService {
  @get
  @route("/admin/admins")
  @Onkt.http
  listAdminAdmins(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/analytics")
  @Onkt.http
  listAdminAnalytics(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/applications/{id}/cv")
  @Onkt.http
  getAdminApplicationsCv(...PathIdRequest): EmptyResponse;

  @get
  @route("/admin/applications/{id}/timeline")
  @Onkt.http
  getAdminApplicationsTimeline(...PathIdRequest): EmptyResponse;

  @get
  @route("/admin/candidates")
  @Onkt.http
  listAdminCandidates(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/cv-jobs")
  @Onkt.http
  listAdminCvJobs(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/leads")
  @Onkt.http
  listAdminLeads(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/me")
  @Onkt.http
  listAdminMe(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/openings")
  @Onkt.http
  listAdminOpenings(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/openings/{id}/applications")
  @Onkt.http
  getAdminOpeningsApplications(...PathIdRequest): EmptyResponse;

  @get
  @route("/admin/shortlists")
  @Onkt.http
  listAdminShortlists(...EmptyRequest): EmptyResponse;

  @get
  @route("/admin/tasks")
  @Onkt.http
  listAdminTasks(...EmptyRequest): EmptyResponse;

  @post
  @route("/admin/admins")
  @Onkt.http
  createAdminAdmins(...EmptyRequest): EmptyResponse;

  @post
  @route("/admin/careers")
  @Onkt.http
  createAdminCareers(...EmptyRequest): EmptyResponse;

  @post
  @route("/admin/login")
  @Onkt.http
  createAdminLogin(...EmptyRequest): EmptyResponse;

  @post
  @route("/admin/shortlists")
  @Onkt.http
  createAdminShortlists(...EmptyRequest): EmptyResponse;

  @post
  @route("/admin/shortlists/{id}/items")
  @Onkt.http
  createAdminShortlistsItems(...PathIdRequest): EmptyResponse;

  @post
  @route("/admin/tasks")
  @Onkt.http
  createAdminTasks(...EmptyRequest): EmptyResponse;

  @post
  @route("/applications")
  @Onkt.http
  @Onkt.maxBody("16MiB")
  createApplications(...EmptyRequest): EmptyResponse;

  @post
  @route("/careers/apply")
  @Onkt.http
  @Onkt.maxBody("16MiB")
  createCareersApply(...EmptyRequest): EmptyResponse;

  @post
  @route("/contact")
  @Onkt.http
  createContact(...EmptyRequest): EmptyResponse;

  @put
  @route("/admin/applications/{id}/review")
  @Onkt.http
  updateAdminApplicationsReview(...PathIdRequest): EmptyResponse;

  @put
  @route("/admin/applications/{id}/status")
  @Onkt.http
  updateAdminApplicationsStatus(...PathIdRequest): EmptyResponse;

  @put
  @route("/admin/careers/{id}")
  @Onkt.http
  updateAdminCareers(...PathIdRequest): EmptyResponse;

  @put
  @route("/admin/careers/{id}/active")
  @Onkt.http
  updateAdminCareersActive(...PathIdRequest): EmptyResponse;

  @put
  @route("/admin/cv-jobs/{source}/{id}/retry")
  @Onkt.http
  updateAdminCvJobsRetry(...PathSourceIdRequest): EmptyResponse;

  @put
  @route("/admin/tasks/{id}")
  @Onkt.http
  updateAdminTasks(...PathIdRequest): EmptyResponse;
}
