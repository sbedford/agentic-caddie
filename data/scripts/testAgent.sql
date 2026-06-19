DELETE FROM shots WHERE hole_id = (559);
DELETE FROM holes WHERE ID = 559;
DELETE FROM rounds WHERE ID=32;

INSERT INTO rounds (id,player_id,course_id,played_at,tees,round_type,competition_type,daily_handicap,total_score,total_points,total_putts,completed,created_at)
VALUES ( 32, 1, 1, '2030-05-24', 'white', 'competition','stableford',10, 4,3,2,FALSE, CURRENT_TIMESTAMP);


INSERT INTO holes (id, round_id, course_hole_id, hole_number, flag_position, score, points, putts, fairway_hit, gir, scramble_save, penalty, completed)
VALUES (559, 32, 1, 1, "back_left", 4,3,2,TRUE,TRUE,null,FALSE,TRUE);

INSERT INTO  shots (id, hole_id, shot_number, shot_type, club, result, miss, strike_quality, pre_shot_recommendation, completed, source)
VALUES
    (2541, 559,1, 'tee', 'driver', 'fairway',null,'clean', '',TRUE,'manual'),
    (2542, 559,2, 'approach', '8i', 'green',null,'clean', '',TRUE,'manual');

-- shot_type values:
--   'tee'       → tee shot on any par (driver, iron, hybrid off the tee)
--   'layup'     → deliberate layup (par 5s, tight holes)
--   'recovery'  → escape from trouble, not going for the green
--   'approach'  → shot intended to reach the green
--   'chip'      → short game shot around the green
--   'bunker'    → bunker shot
--   'putt'      → any shot on the green
--
-- result values:
--   'fairway'   → tee shot on the short grass
--   'rough'     → in the rough
--   'bunker'    → in a bunker
--   'hazard'    → in a penalty area (water, ditch etc)
--   'ob'        → out of bounds
--   'lost'      → lost ball (dense scrub, unplayable and can't find)
--   'green'     → on the putting surface
--   'holed'     → in the hole
--   'unknown'   → reconstructed from spreadsheet, sequence position unclear
--
-- miss values:
--   'left', 'right', 'short', 'long'
--   null if result is clean (fairway, green, holed)
--
-- strike_quality values:
--   'clean', 'thin', 'fat', 'shank'
--   null for reconstructed shots or where not recorded