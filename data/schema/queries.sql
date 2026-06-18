-- ============================================================
-- Golf Caddie - sqlc Query Definitions
-- ============================================================
--
-- One file per logical group, matching the schema table order.
-- All queries use the sqlc annotation format:
--   -- name: QueryName :return_type
--
-- Return types:
--   :one        -> single row, error if not found
--   :many       -> slice of rows
--   :exec       -> no rows returned (INSERT/UPDATE/DELETE)
--   :execresult -> no rows, but returns sql.Result (for LastInsertId)
-- ============================================================


-- ============================================================
-- VOCABULARY
-- Reference data - never inserted or deleted at runtime.
-- GetVocabularyByDomain is the primary call for UI dropdowns and
-- agent pre-flight validation.
-- GetAllVocabulary is the agent bootstrap call to load all enums
-- in one round-trip.
-- ============================================================

-- name: GetVocabularyByDomain :many
SELECT domain, value, label, sort_order
FROM vocabulary
WHERE domain = ?
ORDER BY sort_order;

-- name: GetAllVocabulary :many
SELECT domain, value, label, sort_order
FROM vocabulary
ORDER BY domain, sort_order;

-- name: CreateVocabularyEntry :exec
INSERT INTO vocabulary (domain, value, label, sort_order)
VALUES (?, ?, ?, ?);

-- name: UpdateVocabularyEntry :exec
UPDATE vocabulary
SET label      = ?,
    sort_order = ?
WHERE domain = ?
  AND value  = ?;

-- name: DeleteVocabularyEntry :exec
DELETE FROM vocabulary
WHERE domain = ?
  AND value  = ?;

-- name: VocabValueExists :one
SELECT COUNT(*) FROM vocabulary WHERE domain = ? AND value = ?;


-- ============================================================
-- COURSES
-- Unique constraints: name, golf_api_id
-- ============================================================

-- name: ListCourses :many
SELECT id, name, golf_api_id, created_at
FROM courses
ORDER BY name;

-- name: GetCourseByID :one
SELECT id, name, golf_api_id, created_at
FROM courses
WHERE id = ?;

-- name: GetCourseByName :one
SELECT id, name, golf_api_id, created_at
FROM courses
WHERE name = ?;

-- name: CreateCourse :execresult
INSERT INTO courses (name, golf_api_id, created_at)
VALUES (?, ?, CURRENT_TIMESTAMP);

-- name: UpdateCourse :exec
UPDATE courses
SET name       = ?,
    golf_api_id = ?
WHERE id = ?;

-- name: DeleteCourse :exec
DELETE FROM courses
WHERE id = ?;


-- ============================================================
-- TEES
-- Unique constraint: (course_id, name)
-- ============================================================

-- name: ListTees :many
SELECT id, course_id, name, slope_rating, course_rating
FROM tees
ORDER BY course_id, name;

-- name: GetTeeByID :one
SELECT id, course_id, name, slope_rating, course_rating
FROM tees
WHERE id = ?;

-- name: GetTeesByCourse :many
SELECT id, course_id, name, slope_rating, course_rating
FROM tees
WHERE course_id = ?
ORDER BY name;

-- name: GetTeeByCourseAndName :one
SELECT id, course_id, name, slope_rating, course_rating
FROM tees
WHERE course_id = ?
  AND name = ?;

-- name: CreateTee :execresult
INSERT INTO tees (course_id, name, slope_rating, course_rating)
VALUES (?, ?, ?, ?);

-- name: UpdateTee :exec
UPDATE tees
SET slope_rating  = ?,
    course_rating = ?
WHERE id = ?;

-- name: DeleteTee :exec
DELETE FROM tees
WHERE id = ?;


-- ============================================================
-- COURSE HOLES
-- Unique constraint: (course_id, hole_number)
-- ============================================================

-- name: ListCourseHoles :many
SELECT id, course_id, hole_number, green_centre_lat, green_centre_lng
FROM course_holes
WHERE course_id = ?
ORDER BY hole_number;

-- name: GetCourseHoleByID :one
SELECT id, course_id, hole_number, green_centre_lat, green_centre_lng
FROM course_holes
WHERE id = ?;

-- name: GetCourseHoleByCourseAndNumber :one
SELECT id, course_id, hole_number, green_centre_lat, green_centre_lng
FROM course_holes
WHERE course_id   = ?
  AND hole_number = ?;

-- name: CreateCourseHole :execresult
INSERT INTO course_holes (course_id, hole_number, green_centre_lat, green_centre_lng)
VALUES (?, ?, ?, ?);

-- name: UpdateCourseHoleCoordinates :exec
UPDATE course_holes
SET green_centre_lat = ?,
    green_centre_lng = ?
WHERE id = ?;

-- name: DeleteCourseHole :exec
DELETE FROM course_holes
WHERE id = ?;


-- ============================================================
-- TEE HOLES
-- Unique constraint: (course_hole_id, tee_id)
-- ============================================================

-- name: ListTeeHoles :many
SELECT id, course_hole_id, tee_id, par, stroke_index,
       distance, tee_centre_lat, tee_centre_lng
FROM tee_holes
WHERE tee_id = ?
ORDER BY course_hole_id;

-- name: GetTeeHoleByID :one
SELECT id, course_hole_id, tee_id, par, stroke_index,
       distance, tee_centre_lat, tee_centre_lng
FROM tee_holes
WHERE id = ?;

-- name: GetTeeHoleByHoleAndTee :one
SELECT id, course_hole_id, tee_id, par, stroke_index,
       distance, tee_centre_lat, tee_centre_lng
FROM tee_holes
WHERE course_hole_id = ?
  AND tee_id         = ?;

-- name: GetTeeHoleByCourseIdAndHoleAndTeeName :one
SELECT th.course_hole_id, t.name as tee, th.par, th.stroke_index, th.distance
FROM tee_holes th
INNER JOIN course_holes ch ON ch.id=th.course_hole_id and ch.hole_number=sqlc.arg(HoleNumber)
INNER JOIN tees t ON th.tee_Id = t.id AND t.name = sqlc.arg(TeeName)
INNER JOIN courses c ON c.id=t.course_id and  t.course_id=c.Id AND c.id=sqlc.arg(CourseId);

-- name: GetHolesByCourseAndTee :many
SELECT ch.hole_number, th.Distance, th.Par, th.stroke_index
FROM course_holes ch
INNER JOIN courses c ON c.ID = sqlc.arg(CourseId) AND c.ID = ch.course_id
INNER JOIN tees t ON c.ID=T.course_id AND t.Name=sqlc.arg(TeeName)
INNER JOIN tee_holes th ON th.tee_Id = t.ID AND th.course_hole_id = ch.ID;

-- name: GetHolesByCourse :many
SELECT t.Name as TeeName, ch.hole_number, th.Distance, th.Par, th.stroke_index
FROM course_holes ch
INNER JOIN courses c ON c.ID = sqlc.arg(CourseId) AND c.ID = ch.course_id
INNER JOIN tees t ON c.ID=T.course_id
INNER JOIN tee_holes th ON th.tee_Id = t.ID AND th.course_hole_id = ch.ID;

--CourseId TeeName HoleNumber

-- name: CreateTeeHole :execresult
INSERT INTO tee_holes (
    course_hole_id, tee_id, par, stroke_index,
    distance, tee_centre_lat, tee_centre_lng
)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateTeeHole :exec
UPDATE tee_holes
SET par             = ?,
    stroke_index    = ?,
    distance        = ?,
    tee_centre_lat  = ?,
    tee_centre_lng  = ?
WHERE id = ?;

-- name: DeleteTeeHole :exec
DELETE FROM tee_holes
WHERE id = ?;


-- ============================================================
-- HOLE POINTS OF INTEREST
-- No single-column unique constraint - queried by hole
-- ============================================================

-- name: ListPOIsByHole :many
SELECT id, course_hole_id, specific_tee, poi_type, side,
       reference_point, distance_start, distance_end, label
FROM hole_points_of_interest
WHERE course_hole_id = ?
ORDER BY poi_type, distance_start;

-- name: ListPOIsByHoleAndTee :many
SELECT id, course_hole_id, specific_tee, poi_type, side,
       reference_point, distance_start, distance_end, label
FROM hole_points_of_interest
WHERE course_hole_id = ?
  AND (specific_tee = ? OR specific_tee IS NULL)
ORDER BY poi_type, distance_start;

-- name: GetPOIByID :one
SELECT id, course_hole_id, specific_tee, poi_type, side,
       reference_point, distance_start, distance_end, label
FROM hole_points_of_interest
WHERE id = ?;

-- name: CreatePOI :execresult
INSERT INTO hole_points_of_interest (
    course_hole_id, specific_tee, poi_type, side,
    reference_point, distance_start, distance_end, label
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePOI :exec
UPDATE hole_points_of_interest
SET specific_tee     = ?,
    poi_type         = ?,
    side             = ?,
    reference_point  = ?,
    distance_start   = ?,
    distance_end     = ?,
    label            = ?
WHERE id = ?;

-- name: DeletePOI :exec
DELETE FROM hole_points_of_interest
WHERE id = ?;

-- name: DeletePOIsByHole :exec
DELETE FROM hole_points_of_interest
WHERE course_hole_id = ?;


-- ============================================================
-- PLAYERS
-- Unique constraint: name
-- ============================================================

-- name: ListPlayers :many
SELECT id, name, handicap, updated_at
FROM players
ORDER BY name;

-- name: GetPlayerByID :one
SELECT id, name, handicap, updated_at
FROM players
WHERE id = ?;

-- name: GetPlayerByName :one
SELECT id, name, handicap, updated_at
FROM players
WHERE name = ?;

-- name: CreatePlayer :execresult
INSERT INTO players (name, handicap, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP);

-- name: UpdatePlayerHandicap :exec
UPDATE players
SET handicap   = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeletePlayer :exec
DELETE FROM players
WHERE id = ?;


-- ============================================================
-- PLAYER CLUBS
-- Unique constraint: (player_id, club_name, added_date)
-- Active clubs: removed_date IS NULL
-- ============================================================

-- name: ListClubsByPlayer :many
SELECT id, player_id, club_name, added_date, removed_date,
       carry_avg, carry_reliable, carry_max,
       dispersion_avg_m, dispersion_bias, sample_size, calculated_at
FROM player_clubs
WHERE player_id = ?
ORDER BY added_date DESC, club_name;

-- name: ListActiveClubsByPlayer :many
SELECT id, player_id, club_name, added_date, removed_date,
       carry_avg, carry_reliable, carry_max,
       dispersion_avg_m, dispersion_bias, sample_size, calculated_at
FROM player_clubs
WHERE player_id    = ?
  AND removed_date IS NULL
ORDER BY club_name;

-- name: GetClubByID :one
SELECT id, player_id, club_name, added_date, removed_date,
       carry_avg, carry_reliable, carry_max,
       dispersion_avg_m, dispersion_bias, sample_size, calculated_at
FROM player_clubs
WHERE id = ?;

-- name: GetClubByPlayerAndName :one
-- Returns the current (active) club record for a given player and club name
SELECT id, player_id, club_name, added_date, removed_date,
       carry_avg, carry_reliable, carry_max,
       dispersion_avg_m, dispersion_bias, sample_size, calculated_at
FROM player_clubs
WHERE player_id    = ?
  AND club_name    = ?
  AND removed_date IS NULL;

-- name: GetClubByPlayerNameAndDate :one
-- Returns club as it existed on a specific date (for historical analysis)
SELECT id, player_id, club_name, added_date, removed_date,
       carry_avg, carry_reliable, carry_max,
       dispersion_avg_m, dispersion_bias, sample_size, calculated_at
FROM player_clubs
WHERE player_id  = ?
  AND club_name  = ?
  AND added_date = ?;

-- name: CreateClub :execresult
INSERT INTO player_clubs (
    player_id, club_name, added_date, removed_date,
    carry_avg, carry_reliable, carry_max,
    dispersion_avg_m, dispersion_bias, sample_size, calculated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: RetireClub :exec
-- Marks a club as removed from the bag
UPDATE player_clubs
SET removed_date = ?
WHERE player_id  = ?
  AND club_name  = ?
  AND removed_date IS NULL;

-- name: UpdateClubDistances :exec
-- Updates the distance model - called after each GPS-derived recalculation
UPDATE player_clubs
SET carry_avg        = ?,
    carry_reliable   = ?,
    carry_max        = ?,
    dispersion_avg_m = ?,
    dispersion_bias  = ?,
    sample_size      = ?,
    calculated_at    = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteClub :exec
DELETE FROM player_clubs
WHERE id = ?;


-- ============================================================
-- ROUNDS
-- No single-column unique constraint
-- Primary access patterns: by player, by player+date
-- ============================================================

-- name: ListRounds :many
SELECT id, player_id, course_id, played_at, daily_handicap, tees, round_type, competition_type,
       total_score, total_points, total_putts, created_at
FROM rounds
ORDER BY played_at DESC;

-- name: ListRoundsByPlayer :many
SELECT id, player_id, course_id, played_at, daily_handicap, tees, round_type, competition_type,
       total_score, total_points, total_putts, created_at
FROM rounds
WHERE player_id = ?
ORDER BY played_at DESC;

-- name: ListRoundsByPlayerAndCourse :many
SELECT id, player_id, course_id, played_at, daily_handicap, tees, round_type, competition_type,
       total_score, total_points, total_putts, created_at
FROM rounds
WHERE player_id = ?
  AND course_id = ?
ORDER BY played_at DESC;

-- name: GetRoundByID :one
SELECT id, player_id, course_id, played_at, daily_handicap, tees, round_type, competition_type,
       total_score, total_points, total_putts, created_at
FROM rounds
WHERE id = ?;

-- name: GetRoundByPlayerAndDate :one
SELECT id, player_id, course_id, played_at, daily_handicap, tees, round_type, competition_type,
       total_score, total_points, total_putts, created_at
FROM rounds
WHERE player_id = ?
  AND played_at = ?;

-- name: CreateRound :execresult
INSERT INTO rounds (
    player_id, course_id, played_at, tees, daily_handicap, round_type, competition_type,
    total_score, total_points, total_putts, created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: UpdateRoundTotals :exec
-- Called after holes are inserted/updated to refresh denormalised totals
UPDATE rounds
SET total_score  = (SELECT SUM(score)  FROM holes WHERE round_id = rounds.id),
    total_points = (SELECT SUM(points) FROM holes WHERE round_id = rounds.id),
    total_putts  = (SELECT SUM(putts)  FROM holes WHERE round_id = rounds.id)
WHERE rounds.id = ?;

-- name: DeleteRound :exec
DELETE FROM rounds
WHERE id = ?;


-- ============================================================
-- HOLES
-- Unique constraint: (round_id, hole_number)
-- ============================================================

-- name: ListHolesByRound :many
SELECT id, round_id, course_hole_id, hole_number, flag_position,
       score, points, putts,fairway_hit,  gir,  scramble_save, penalty
FROM holes
WHERE round_id = ?
ORDER BY hole_number;

-- name: GetHoleByID :one
SELECT id, round_id, course_hole_id, hole_number, flag_position,
       score, points, putts,fairway_hit,  gir, scramble_save, penalty
FROM holes
WHERE id = ?;

-- name: GetHoleByRoundAndNumber :one
SELECT id, round_id, course_hole_id, hole_number, flag_position,
       score, points, putts, fairway_hit,  gir, scramble_save, penalty
FROM holes
WHERE round_id    = ?
  AND hole_number = ?;

-- name: CreateHole :execresult
INSERT INTO holes (
    round_id, course_hole_id, hole_number, flag_position,
    score, points, putts, gir, scramble_save, penalty
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateHole :exec
UPDATE holes
SET flag_position  = ?,
    score          = ?,
    points         = ?,
    putts          = ?,
    gir            = ?,
    scramble_save  = ?,
    penalty        = ?
WHERE id = ?;

-- name: DeleteHole :exec
DELETE FROM holes
WHERE id = ?;


-- ============================================================
-- SHOTS
-- Unique constraint: (hole_id, shot_number)
-- ============================================================

-- name: ListShotsByHole :many
SELECT id, hole_id, shot_number, shot_type, club,
       result, miss, strike_quality, source
FROM shots
WHERE hole_id = ?
ORDER BY shot_number;

-- name: ListShotsByHoleAndType :many
SELECT id, hole_id, shot_number, shot_type, club,
       result, miss, strike_quality, source
FROM shots
WHERE hole_id   = ?
  AND shot_type = ?
ORDER BY shot_number;

-- name: GetShotByID :one
SELECT id, hole_id, shot_number, shot_type, club,
       result, miss, strike_quality, source
FROM shots
WHERE id = ?;

-- name: GetShotByHoleAndNumber :one
SELECT id, hole_id, shot_number, shot_type, club,
       result, miss, strike_quality, source
FROM shots
WHERE hole_id    = ?
  AND shot_number = ?;

-- name: CreateShot :execresult
INSERT INTO shots (
    hole_id, shot_number, shot_type, club,
    result, miss, strike_quality, source
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateShot :exec
UPDATE shots
SET shot_type     = ?,
    club          = ?,
    result        = ?,
    miss          = ?,
    strike_quality = ?,
    source        = ?
WHERE id = ?;

-- name: DeleteShot :exec
DELETE FROM shots
WHERE id = ?;

-- name: DeleteShotsByHole :exec
DELETE FROM shots
WHERE hole_id = ?;


-- ============================================================
-- COMMENTARY
-- No unique constraint - scope + scope_id can have multiple entries
-- (multiple commentary passes on the same hole/round are valid)
-- ============================================================

-- name: ListCommentaryByScope :many
SELECT id, scope, scope_id, content, generated_at
FROM commentary
WHERE scope    = ?
  AND scope_id = ?
ORDER BY generated_at DESC;

-- name: GetCommentaryByID :one
SELECT id, scope, scope_id, content, generated_at
FROM commentary
WHERE id = ?;

-- name: GetLatestCommentaryByScope :one
-- Returns the most recent commentary for a hole or round
SELECT id, scope, scope_id, content, generated_at
FROM commentary
WHERE scope    = ?
  AND scope_id = ?
ORDER BY generated_at DESC
LIMIT 1;

-- name: CreateCommentary :execresult
INSERT INTO commentary (scope, scope_id, content, generated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP);

-- name: DeleteCommentaryByScope :exec
DELETE FROM commentary
WHERE scope    = ?
  AND scope_id = ?;

-- name: DeleteCommentary :exec
DELETE FROM commentary
WHERE id = ?;


-- ============================================================
-- Statistics
-- Used by agent tools to summarise performance across different dimensions
-- ============================================================

-- name: GetHoleStats :many
SELECT r.played_at, h.score, h.points, h.putts, h.fairway_hit, h.gir
FROM holes h 
INNER JOIN course_holes ch ON h.course_hole_id=ch.id and ch.course_id = sqlc.arg(CourseId)
INNER JOIN rounds r ON r.id = h.round_id
INNER JOIN tees t ON t.course_id = ch.course_id and t.name = sqlc.arg(TeeName)
WHERE h.hole_number=sqlc.arg(HoleNumber)
ORDER BY r.played_at DESC;

