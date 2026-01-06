-- Add 'no_house' option to house_type enum
ALTER TYPE house_type ADD VALUE IF NOT EXISTS 'no_house';
