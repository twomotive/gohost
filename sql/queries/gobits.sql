-- name: CreateGobit :one
INSERT INTO gobits (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;


-- name: GetAllGobits :many
SELECT * FROM gobits
ORDER BY created_at ASC;



-- name: GetGobit :one
SELECT * FROM gobits WHERE id = $1;