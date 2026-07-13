.PHONY: clean lint format breaking generate install-plugins install-onek install-js typecheck

ONEKIT_VERSION := v0.3.5
ONK_VERSION := v0.4.4

# Proto paths already migrated to .onk (see api/) - their proto files are
# intentionally deleted, not accidentally regressed, so they're excluded from
# breaking-change detection. Add the migrated service's proto path here as
# part of each migration PR (see README.md "Migrating a service to .onk").
#
# common/pagination/v1 is NOT in this list: it's a shared shim still imported
# by unmigrated insights/report/v1, insights/customer/v1, so its proto/ files
# stay in place (alongside the new api/common/pagination/v1) until those are
# migrated too.
BREAKING_EXCLUDE := \
  --exclude-path proto/hub/business/v1 \
  --exclude-path proto/hub/config/v1 \
  --exclude-path proto/booking/appointment/v1 \
  --exclude-path proto/booking/calendar/v1 \
  --exclude-path proto/booking/dashboard/v1 \
  --exclude-path proto/booking/reminder/v1 \
  --exclude-path proto/booking/resource/v1 \
  --exclude-path proto/booking/service/v1 \
  --exclude-path proto/booking/waitlist/v1 \
  --exclude-path proto/common/auth/v1 \
  --exclude-path proto/common/oauth/v1 \
  --exclude-path proto/console/admin/v1 \
  --exclude-path proto/support/canned/v1 \
  --exclude-path proto/support/customer/v1 \
  --exclude-path proto/support/inbox/v1 \
  --exclude-path proto/support/knowledge/v1 \
  --exclude-path proto/support/settings/v1 \
  --exclude-path proto/erp/brand/v1 \
  --exclude-path proto/erp/category/v1 \
  --exclude-path proto/erp/inventory/v1 \
  --exclude-path proto/erp/price_list/v1 \
  --exclude-path proto/erp/product/v1 \
  --exclude-path proto/erp/purchase_order/v1 \
  --exclude-path proto/erp/stock_take/v1 \
  --exclude-path proto/erp/stock_transfer/v1 \
  --exclude-path proto/erp/supplier/v1 \
  --exclude-path proto/erp/variant/v1 \
  --exclude-path proto/erp/warehouse/v1 \
  --exclude-path proto/crm/campaign/v1 \
  --exclude-path proto/crm/customer/v1 \
  --exclude-path proto/crm/dashboard/v1 \
  --exclude-path proto/crm/feedback/v1 \
  --exclude-path proto/crm/loyalty/v1 \
  --exclude-path proto/crm/segment/v1 \
  --exclude-path proto/crm/settings/v1 \
  --exclude-path proto/crm/template/v1 \
  --exclude-path proto/finance/account/v1 \
  --exclude-path proto/finance/bank_account/v1 \
  --exclude-path proto/finance/bill/v1 \
  --exclude-path proto/finance/dashboard/v1 \
  --exclude-path proto/finance/expenses/v1 \
  --exclude-path proto/finance/invoices/v1 \
  --exclude-path proto/finance/invoices/payments/v1 \
  --exclude-path proto/finance/journal/v1 \
  --exclude-path proto/finance/quote/v1 \
  --exclude-path proto/finance/reconciliation/v1 \
  --exclude-path proto/finance/recurring_invoice/v1 \
  --exclude-path proto/finance/report/v1 \
  --exclude-path proto/finance/statement/v1 \
  --exclude-path proto/finance/tax/v1 \
  --exclude-path proto/flow/cycle/v1 \
  --exclude-path proto/flow/dashboard/v1 \
  --exclude-path proto/flow/insights/v1 \
  --exclude-path proto/flow/label/v1 \
  --exclude-path proto/flow/milestone/v1 \
  --exclude-path proto/flow/notification/v1 \
  --exclude-path proto/flow/project/v1 \
  --exclude-path proto/flow/setting/v1 \
  --exclude-path proto/flow/task/v1 \
  --exclude-path proto/flow/view/v1 \
  --exclude-path proto/flow/workspace/v1

# Clean generated code
clean:
	rm -rf gen

# Lint proto files
lint:
	buf lint

# Format proto files
format:
	buf format -w

# Check for breaking changes
breaking:
	buf breaking --against '.git#branch=main' $(BREAKING_EXCLUDE)

# Install the onekit protoc plugins used for services still on .proto (see proto/).
# Being migrated off in favor of onek/.onk - see api/ and README.md.
install-plugins:
	go install github.com/corezio/onekit/cmd/protoc-gen-onekit-go-http@$(ONEKIT_VERSION)
	go install github.com/corezio/onekit/cmd/protoc-gen-onekit-ts-client@$(ONEKIT_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install the onek CLI used for services migrated to .onk (see api/).
install-onek:
	go install github.com/1homsi/onekit/cmd/onek@$(ONK_VERSION)

# Install JS codegen tooling (ts-proto)
install-js:
	pnpm install --frozen-lockfile

# Regenerate gen/go and gen/ts from ALL schema sources: proto/ (buf, being
# migrated off) and api/ (onek, the target). Output is committed to the repo.
generate: install-plugins install-onek install-js
	buf dep update
	rm -rf gen/go gen/ts
	buf generate
	python3 scripts/generate_validation_metadata.py --proto-root proto --out-dir gen/ts/validation
	onek build .

# Typecheck the generated TypeScript output
typecheck: install-js
	pnpm exec tsc --noEmit
