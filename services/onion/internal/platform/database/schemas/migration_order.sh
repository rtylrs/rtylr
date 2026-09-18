#!/usr/bin/env bash

# Keep every schema entry point on the same dependency order.  The onion
# repository sources this file from scripts/migrate.sh, while this repository's
# runner sources it directly from migrate.sh.
SCHEMA_MIGRATION_DIRS=(
    shared
    cdn
    rtylr
    voxire
    erp
    hr
    hr_core
    hr_attendance
    hr_documents
    hr_equipment
    hr_leave
    hr_onboarding
    hr_overtime
    hr_payroll
    hr_performance
    hr_setup
    hr_shifts
    hr_time
    crm
    crm_migrations
    finance
    flow
    pos
    hub
    store
    support
    book
    hr_dashboard
    recruit
    insights
    oauth
    ai
    migrations
    _perf
)
