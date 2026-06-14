-- name: ListPlayers :many
SELECT * 
FROM players;

-- name: GetPlayerByName :one
SELECT * 
from players p 
WHERE p.name=?;

-- name: GetPlayerById :one
SELECT * from players WHERE id = ?;

-- name: GetClubDistances :many
SELECT p.Id, p.Name, pc.club_name, pc.carry_avg, pc.carry_reliable, pc.carry_max, pc.dispersion_avg_m, pc.dispersion_bias 
FROM players p 
    INNER JOIN player_clubs pc ON p.id = pc.player_id 
WHERE p.Name = ?;

-- name: GetCourseByName :one
SELECT * FROM courses WHERE name = ?;

-- name: GetTeesByCourse :many
SELECT ct.id, ct.name, ct.slope_rating, ct.course_rating FROM tees ct WHERE course_id=?;