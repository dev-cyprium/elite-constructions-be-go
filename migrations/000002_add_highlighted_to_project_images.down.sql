-- Remove index
DROP INDEX IF EXISTS idx_project_images_highlighted;

-- Remove highlighted column from project_images table
ALTER TABLE project_images DROP COLUMN IF EXISTS highlighted;
