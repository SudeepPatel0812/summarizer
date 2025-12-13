-- Create table for storing relevant information about
-- videos.

CREATE TABLE videos (
    name NVARCHAR(255),
    description NVARCHAR(255)
    length NVARCHAR(255),
    author NVARCHAR(255),
    upload_date DATETIME,
    added_date DATETIME,
    updated_date DATETIME
)

-- Create table for storing the summarized text about
-- videos.
CREATE TABLE summarized_video(
    
)