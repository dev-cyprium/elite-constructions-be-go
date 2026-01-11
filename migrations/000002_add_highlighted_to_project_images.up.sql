-- Add highlighted column to project_images table
ALTER TABLE project_images ADD COLUMN highlighted BOOLEAN NOT NULL DEFAULT false;

-- Create index for better query performance
CREATE INDEX idx_project_images_highlighted ON project_images(project_id, highlighted) WHERE highlighted = true;
