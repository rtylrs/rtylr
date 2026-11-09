//! Unit tests for the sqlc2sqlx parser.

#[cfg(test)]
mod tests {
    use super::{extract_params_struct, named_binds, snake};

    #[test]
    fn params_struct_keeps_every_field() {
        let source = "type ReserveMonthlyTokensParams struct {\n\tEstimate     int64       `db:\"estimate\" json:\"estimate\"`\n\tBusinessID   int64       `db:\"business_id\" json:\"business_id\"`\n\tUserID       int64       `db:\"user_id\" json:\"user_id\"`\n\tBillingMonth time.Time   `db:\"billing_month\" json:\"billing_month\"`\n\tAllowOverage interface{} `db:\"allow_overage\" json:\"allow_overage\"`\n\tAllowance    int64       `db:\"allowance\" json:\"allowance\"`\n}";
        let fields = extract_params_struct(source, "ReserveMonthlyTokensParams");
        assert_eq!(fields.len(), 6, "{fields:?}");
        assert_eq!(fields[5].db, "allowance");
    }

    #[test]
    fn named_binds_find_every_occurrence() {
        let sql = "UPDATE x SET a = sqlc.arg('estimate') WHERE b = sqlc.arg('business_id') AND c = sqlc.arg('estimate') OR d = sqlc.arg('allowance')";
        let binds = named_binds(sql);
        assert_eq!(
            binds,
            vec!["estimate", "business_id", "estimate", "allowance"],
            "{binds:?}"
        );
    }

    #[test]
    fn snake_handles_acronyms() {
        assert_eq!(snake("BusinessID"), "business_id");
        assert_eq!(snake("AISlug"), "ai_slug");
        assert_eq!(snake("UserID"), "user_id");
    }
}
