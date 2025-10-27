-- Create posts table
CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    caption TEXT NOT NULL,
    image_path VARCHAR(500) NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    creator_id BIGINT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    creator_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE NULL
);

-- Create indexes for posts
CREATE INDEX IF NOT EXISTS idx_posts_creator_id ON posts (creator_id);

CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts (deleted_at);

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    post_id BIGINT NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    creator_id BIGINT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    creator_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE NULL
);

-- Create indexes for comments
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id);

CREATE INDEX IF NOT EXISTS idx_comments_creator_id ON comments (creator_id);

CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments (deleted_at);

-- Create a view for posts with comment counts (for sorting by comment count)
CREATE OR REPLACE VIEW posts_with_comment_count AS
SELECT p.*, COALESCE(
        comment_counts.comment_count, 0
    ) as comment_count
FROM posts p
    LEFT JOIN (
        SELECT post_id, COUNT(*) as comment_count
        FROM comments
        WHERE
            deleted_at IS NULL
        GROUP BY
            post_id
    ) comment_counts ON p.id = comment_counts.post_id
WHERE
    p.deleted_at IS NULL;