-- name: ListTrainingCourses :many
SELECT id, owner_id, name, description, duration_hours, is_required, created_at
FROM hr_training_courses
WHERE owner_id = ?
ORDER BY name ASC;

-- name: CreateTrainingCourse :execresult
INSERT INTO hr_training_courses (owner_id, name, description, duration_hours, is_required)
VALUES (?, ?, ?, ?, ?);
