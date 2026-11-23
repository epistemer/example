-- Users (the auth Directory)

-- name: GetUser :one
SELECT * FROM staff WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM staff WHERE lower(email) = lower($1);

-- name: UpsertUser :exec
INSERT INTO staff (id, email, name) VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name;

-- Events

-- name: GetEvent :one
SELECT * FROM event WHERE id = $1;

-- name: GetEventBySlug :one
SELECT * FROM event WHERE slug = $1;

-- name: CreateEvent :one
INSERT INTO event (name, slug, summary, status, starts_on, ends_on, online, capacity,
    venue_name, venue_street, venue_city, venue_country, venue_post_code,
    latitude, longitude, currency)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING id;

-- name: UpdateEvent :execrows
UPDATE event SET name = $2, slug = $3, summary = $4, status = $5, starts_on = $6,
    ends_on = $7, online = $8, capacity = $9, venue_name = $10, venue_street = $11,
    venue_city = $12, venue_country = $13, venue_post_code = $14, latitude = $15,
    longitude = $16, currency = $17, version = version + 1, updated_at = now()
WHERE id = $1 AND version = $18;

-- name: SetEventStatus :exec
UPDATE event SET status = $2, version = version + 1, updated_at = now() WHERE id = $1;

-- name: DeleteEvent :exec
DELETE FROM event WHERE id = $1;

-- name: ListEventOptions :many
SELECT id, name FROM event ORDER BY starts_on DESC, id;

-- name: ListTicketTiers :many
SELECT * FROM ticket_tier WHERE event_id = $1 ORDER BY position;

-- name: DeleteTicketTiers :exec
DELETE FROM ticket_tier WHERE event_id = $1;

-- name: CreateTicketTier :exec
INSERT INTO ticket_tier (event_id, position, name, price, quantity, sale_ends_on)
VALUES ($1, $2, $3, $4, $5, $6);

-- Speakers

-- name: GetSpeaker :one
SELECT * FROM speaker WHERE id = $1;

-- name: SearchSpeakers :many
SELECT id, name, email FROM speaker
WHERE name ILIKE '%' || sqlc.arg(q)::text || '%' OR email ILIKE '%' || sqlc.arg(q)::text || '%'
ORDER BY name, id LIMIT sqlc.arg(lim);

-- name: CreateSpeaker :one
INSERT INTO speaker (name, email, phone, country, website, bio, photo_key, photo_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: UpdateSpeaker :exec
UPDATE speaker SET name = $2, email = $3, phone = $4, country = $5, website = $6,
    bio = $7, photo_key = $8, photo_name = $9
WHERE id = $1;

-- name: DeleteSpeaker :exec
DELETE FROM speaker WHERE id = $1;

-- Talks

-- name: GetTalk :one
SELECT * FROM talk WHERE id = $1;

-- name: GetTalkView :one
SELECT * FROM talk_view WHERE id = $1;

-- name: CreateTalk :one
INSERT INTO talk (event_id, speaker_id, title, abstract, level, format, topics,
    starts_at, duration_minutes, seats, recorded)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: UpdateTalk :exec
UPDATE talk SET event_id = $2, speaker_id = $3, title = $4, abstract = $5, level = $6,
    format = $7, topics = $8, starts_at = $9, duration_minutes = $10, seats = $11,
    recorded = $12, updated_at = now()
WHERE id = $1;

-- name: SetTalkState :exec
UPDATE talk SET state = $2, updated_at = now() WHERE id = $1;

-- name: SetTalksState :execrows
UPDATE talk SET state = sqlc.arg(state), updated_at = now() WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: DeleteTalk :exec
DELETE FROM talk WHERE id = $1;

-- name: OverlappingTalk :one
SELECT id, title FROM talk
WHERE event_id = sqlc.arg(event_id) AND speaker_id = sqlc.arg(speaker_id) AND id <> sqlc.arg(id) AND starts_at IS NOT NULL
  AND starts_at < sqlc.arg(ends_at)::timestamptz
  AND starts_at + make_interval(mins => duration_minutes) > sqlc.arg(starts_at)::timestamptz
LIMIT 1;

-- name: PublicSchedule :many
SELECT t.id, t.title, t.level, t.format, t.starts_at, t.duration_minutes, s.name AS speaker
FROM talk t JOIN speaker s ON s.id = t.speaker_id
WHERE t.event_id = $1 AND t.state = 'scheduled'
ORDER BY t.starts_at, t.id;

-- Talk files (slides)

-- name: ListTalkFiles :many
SELECT * FROM talk_file WHERE talk_id = $1 ORDER BY id;

-- name: GetTalkFileByKey :one
SELECT * FROM talk_file WHERE key = $1;

-- name: CreateTalkFile :exec
INSERT INTO talk_file (talk_id, key, name, content_type, size) VALUES ($1, $2, $3, $4, $5);

-- name: DeleteTalkFile :exec
DELETE FROM talk_file WHERE talk_id = $1 AND key = $2;
