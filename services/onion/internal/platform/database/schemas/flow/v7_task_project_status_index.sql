ALTER TABLE flow_tasks
  ADD INDEX idx_flow_tasks_project_status_parent_order (
    project_id,
    status,
    parent_task_id,
    sort_order,
    created_at
  );
