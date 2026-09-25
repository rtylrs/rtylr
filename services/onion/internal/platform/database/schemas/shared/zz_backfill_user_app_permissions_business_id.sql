UPDATE user_app_permissions p
JOIN businesses b ON b.owner_id = p.owner_id
SET p.business_id = b.id
WHERE p.business_id IS NULL;
