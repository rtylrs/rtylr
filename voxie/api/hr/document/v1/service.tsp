import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.document.v1;

@route("/hr/document/v1")
interface DocumentService {
  @get
  @route("/documents")
  listDocuments(...ListDocumentsRequest): ListDocumentsResponse;

  @post
  @route("/documents")
  createDocument(...CreateDocumentRequest): CreateDocumentResponse;

  @delete
  @route("/documents/{id}")
  deleteDocument(...DeleteDocumentRequest): DeleteDocumentResponse;
}
