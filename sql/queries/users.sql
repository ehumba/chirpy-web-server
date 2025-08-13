-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: LookUpByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: LookUpByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserData :exec
UPDATE users
SET email = $2,
hashed_password = $3,
updated_at = NOW()
WHERE id = $1;


-- name: MakeChirpyRed :exec
UPDATE users
SET is_chirpy_red = TRUE,
updated_at = NOW()
WHERE id = $1
AND is_chirpy_red = FALSE;