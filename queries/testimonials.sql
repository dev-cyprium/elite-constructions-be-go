-- name: ListTestimonials :many
SELECT * FROM testimonials ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountTestimonials :one
SELECT COUNT(*) FROM testimonials;

-- name: ListPublicTestimonials :many
SELECT * FROM testimonials WHERE status = 'ready' ORDER BY created_at DESC;

-- name: GetTestimonialByID :one
SELECT * FROM testimonials WHERE id = $1;

-- name: CreateTestimonial :one
INSERT INTO testimonials (full_name, profession, testimonial, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING *;

-- name: UpdateTestimonial :exec
UPDATE testimonials
SET full_name = $2,
    profession = $3,
    testimonial = $4,
    status = $5,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteTestimonial :exec
DELETE FROM testimonials WHERE id = $1;
