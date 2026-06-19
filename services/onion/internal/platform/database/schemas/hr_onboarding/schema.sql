CREATE TABLE IF NOT EXISTS hr_onboarding_checklist_templates (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'onboarding',
    role_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_template_name (owner_id, name),
    INDEX idx_onboarding_templates_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_onboarding_checklist_items (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    template_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignee_role VARCHAR(100) NOT NULL DEFAULT 'employee',
    due_offset_days INT NOT NULL DEFAULT 0,
    requires_attachment BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_onboarding_items_owner (owner_id),
    INDEX idx_onboarding_items_template (owner_id, template_id)
);

CREATE TABLE IF NOT EXISTS hr_onboarding_employee_checklists (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    template_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    triggered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_onboarding_emp_checklists_owner (owner_id),
    INDEX idx_onboarding_emp_checklists_employee (owner_id, employee_id)
);

CREATE TABLE IF NOT EXISTS hr_onboarding_employee_checklist_items (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_checklist_id BIGINT NOT NULL,
    checklist_item_id BIGINT NOT NULL,
    assigned_to_employee_id BIGINT NOT NULL DEFAULT 0,
    due_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'incomplete',
    completed_at TIMESTAMP,
    completed_by BIGINT,
    attachment_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_onboarding_emp_items_owner (owner_id),
    INDEX idx_onboarding_emp_items_checklist (owner_id, employee_checklist_id)
);
