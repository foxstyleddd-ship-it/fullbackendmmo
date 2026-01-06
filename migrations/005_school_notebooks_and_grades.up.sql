-- 005: School Notebooks and Academic Grades System

-- Create subject enum (matières d'enseignement)
CREATE TYPE school_subject AS ENUM (
    'charms',
    'transfiguration',
    'potions',
    'defense_against_dark_arts',
    'herbology',
    'astronomy',
    'history_of_magic',
    'care_of_magical_creatures',
    'divination',
    'ancient_runes',
    'arithmancy',
    'muggle_studies',
    'flying'
);

-- Create grade value enum (mentions)
CREATE TYPE grade_value AS ENUM (
    'outstanding',      -- Optimal (O)
    'exceeds_expectations', -- Effort Exceptionnel (E)
    'acceptable',       -- Acceptable (A)
    'poor',            -- Piètre (P)
    'dreadful',        -- Désolant (D)
    'troll'            -- Troll (T)
);

-- Table: school_notebooks (Carnets d'école)
CREATE TABLE school_notebooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    title VARCHAR(128) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    subject school_subject,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT notebooks_character_fk FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

-- Indexes for school_notebooks
CREATE INDEX idx_notebooks_character ON school_notebooks(character_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notebooks_subject ON school_notebooks(subject) WHERE deleted_at IS NULL;

-- Table: academic_grades (Notes par matière par année)
CREATE TABLE academic_grades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    school_year INTEGER NOT NULL, -- Année d'étude (1-7)
    subject school_subject NOT NULL,
    grade_value grade_value NOT NULL,
    exam_type VARCHAR(32) NOT NULL DEFAULT 'continuous', -- continuous, midterm, final, owls, newts
    score INTEGER, -- Score numérique optionnel (0-100)
    teacher_comment TEXT,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT grades_character_fk FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT grades_year_check CHECK (school_year >= 1 AND school_year <= 7),
    CONSTRAINT grades_score_check CHECK (score IS NULL OR (score >= 0 AND score <= 100))
);

-- Indexes for academic_grades
CREATE INDEX idx_grades_character ON academic_grades(character_id);
CREATE INDEX idx_grades_year ON academic_grades(school_year);
CREATE INDEX idx_grades_subject ON academic_grades(subject);
CREATE INDEX idx_grades_character_year ON academic_grades(character_id, school_year);

-- Table: report_cards (Bulletins de fin d'année)
CREATE TABLE report_cards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    school_year INTEGER NOT NULL,
    overall_comment TEXT, -- Remarque globale
    headmaster_signature VARCHAR(128),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT report_cards_character_fk FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT report_cards_year_check CHECK (school_year >= 1 AND school_year <= 7),
    CONSTRAINT report_cards_unique UNIQUE (character_id, school_year)
);

-- Indexes for report_cards
CREATE INDEX idx_report_cards_character ON report_cards(character_id);
CREATE INDEX idx_report_cards_year ON report_cards(school_year);

-- Trigger to update updated_at for school_notebooks
CREATE TRIGGER update_notebooks_updated_at
    BEFORE UPDATE ON school_notebooks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Trigger to update updated_at for academic_grades
CREATE TRIGGER update_grades_updated_at
    BEFORE UPDATE ON academic_grades
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Insert some example data for testing
-- (This would typically be done through the API, but we can seed some initial structure)

COMMENT ON TABLE school_notebooks IS 'Personal notebooks where students can take notes during classes';
COMMENT ON TABLE academic_grades IS 'Academic grades per subject per year, including exams (OWLs, NEWTs)';
COMMENT ON TABLE report_cards IS 'End-of-year report cards with overall comments';
COMMENT ON TYPE school_subject IS 'School subjects taught at Hogwarts';
COMMENT ON TYPE grade_value IS 'Grade values: Outstanding, Exceeds Expectations, Acceptable, Poor, Dreadful, Troll';
