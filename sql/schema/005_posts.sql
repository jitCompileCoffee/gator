-- +goose Up
CREATE TABLE posts (
	id				UUID PRIMARY KEY,
	created_at		TIMESTAMP NOT NULL,
	updated_at		TIMESTAMP NOT NULL,
	title			TEXT NOT NULL,
	url				TEXT NOT NULL,
	description		TEXT,
	published_at	TIMESTAMP,
	feed_id			UUID NOT NULL,
	CONSTRAINT fk_feeds
		FOREIGN KEY (feed_id)
		REFERENCES feeds(id)
		ON DELETE CASCADE,
	CONSTRAINT uq_posts_url UNIQUE (url)
);

-- +goose Down
DROP TABLE posts;
