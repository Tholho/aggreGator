-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE name = $1;

-- name: GetFeeds :many
SELECT f.name as feedname, f.url, u.name as username
FROM users u
JOIN feeds f ON f.user_id = u.id;

-- name: GetUsers :many
 SELECT * FROM users;


-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (
INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT ff.*, f.name as feedname, u.name as username
FROM inserted_feed_follow AS ff
INNER JOIN users u ON ff.user_id = u.id
INNER JOIN feeds f ON ff.feed_id = f.id;

-- name: GetFeedByURL :one
SELECT f.name as feedname, f.id
FROM feeds f
WHERE f.url = $1;

-- name: GetFeedFollowsForUser :many
WITH userID AS (
    SELECT id
    FROM users
    WHERE users.name = $1
)
SELECT feeds.name
FROM feed_follows
JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = (SELECT id FROM userID);

-- name: DeleteFeedFollow :exec
WITH target_feed_id AS (
    SELECT id
    FROM feeds
    WHERE url = $2
)
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1
AND feed_id = (SELECT id from target_feed_id);