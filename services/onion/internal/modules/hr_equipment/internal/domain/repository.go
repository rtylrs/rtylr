package domain

import "context"

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}
