-- Rollback: Remove school notebooks and academic grades system

-- Drop triggers
DROP TRIGGER IF EXISTS update_notebooks_updated_at ON school_notebooks;
DROP TRIGGER IF EXISTS update_grades_updated_at ON academic_grades;

-- Drop tables
DROP TABLE IF EXISTS report_cards;
DROP TABLE IF EXISTS academic_grades;
DROP TABLE IF EXISTS school_notebooks;

-- Drop enum types
DROP TYPE IF EXISTS grade_value;
DROP TYPE IF EXISTS school_subject;
