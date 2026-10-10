package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, service *application.Service) {
	serverHttp.RegisterPOSTWithValidation[UploadDocumentRequest, UploadDocumentResponse](
		ctx, router, UploadDocumentRoutes(), &UploadDocumentHandler{Service: service},
	)
	serverHttp.RegisterGET[ListDocumentsRequest, any](
		ctx, router, ListDocumentsRoutes(), &ListDocumentsHandler{Service: service},
	)
	serverHttp.RegisterGET[ListExpiringDocumentsRequest, any](
		ctx, router, ListExpiringDocumentsRoutes(), &ListExpiringDocumentsHandler{Service: service},
	)
	serverHttp.RegisterGET[GetDocumentRequest, GetDocumentResponse](
		ctx, router, GetDocumentRoutes(), &GetDocumentHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteDocumentRequest, DeleteDocumentResponse](
		ctx, router, DeleteDocumentRoutes(), &DeleteDocumentHandler{Service: service},
	)
}
