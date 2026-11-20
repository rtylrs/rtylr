import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.template.v1;

@route("/crm/template/v1")
interface TemplateService {
  @get
  @route("/templates")
  listTemplates(...ListTemplatesRequest): ListTemplatesResponse;

  @post
  @route("/templates")
  createTemplate(...CreateTemplateRequest): CreateTemplateResponse;

  @put
  @route("/templates/{id}")
  updateTemplate(...UpdateTemplateRequest): UpdateTemplateResponse;

  @delete
  @route("/templates/{id}")
  deleteTemplate(...DeleteTemplateRequest): DeleteTemplateResponse;
}
