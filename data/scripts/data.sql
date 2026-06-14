-- --------------------------------------------------------
-- Seed: Player and Club Distances
-- Player: Sean Bedford
-- Generated: 2026-06-14
--
-- Notes:
--   - player_clubs records the bag inventory (what clubs exist)
--   - player_club_distances seeds the distance model manually
--   - sample_size = 0 indicates manually seeded, not GPS-derived
--   - carry_avg and carry_max are null until GPS data is available
--   - dispersion_bias uses lowercase to match schema enum convention
-- --------------------------------------------------------


-- Player
-- Assumes no existing player record. Adjust id if inserting into
-- a database that already has rows in the players table.

INSERT INTO players (id, name, handicap, updated_at)
VALUES (1, 'Sean Bedford', 8.9, CURRENT_TIMESTAMP);


-- --------------------------------------------------------
-- player_clubs
-- Records what is currently in the bag (no removed_date)
-- added_date set to seed date as a reasonable baseline
-- --------------------------------------------------------

INSERT INTO player_clubs (player_id, club_name, added_date) VALUES
    (1, 'driver',    '2026-06-14'),
    (1, '3_hybrid',  '2026-06-14'),
    (1, '3i',        '2026-06-14'),
    (1, '4i',        '2026-06-14'),
    (1, '5i',        '2026-06-14'),
    (1, '6i',        '2026-06-14'),
    (1, '7i',        '2026-06-14'),
    (1, '8i',        '2026-06-14'),
    (1, '9i',        '2026-06-14'),
    (1, 'pw',        '2026-06-14'),
    (1, 'gw',        '2026-06-14'),
    (1, 'sw',        '2026-06-14');


-- --------------------------------------------------------
-- player_club_distances
-- Manually seeded reliable carry distances from CSV.
-- carry_avg and carry_max are null — not yet GPS-derived.
-- dispersion_avg_m and dispersion_bias are best estimates.
-- sample_size = 0 flags these as manual seeds throughout.
-- --------------------------------------------------------

INSERT INTO player_club_distances (
    player_id,
    club_name,
    carry_avg,
    carry_reliable,
    carry_max,
    dispersion_avg_m,
    dispersion_bias,
    sample_size,
    calculated_at
) VALUES
    (1, 'driver',   NULL, 220, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '3_hybrid', NULL, 200, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '3i',       NULL, 190, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '4i',       NULL, 180, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '5i',       NULL, 170, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '6i',       NULL, 160, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '7i',       NULL, 150, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '8i',       NULL, 140, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, '9i',       NULL, 130, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, 'pw',       NULL, 115, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, 'gw',       NULL, 100, NULL, 20, 'left', 0, CURRENT_TIMESTAMP),
    (1, 'sw',       NULL,  85, NULL, 20, 'left', 0, CURRENT_TIMESTAMP);

    -- --------------------------------------------------------
-- Seed: St Michaels Golf Course
-- Tees: White (Slope 133, Rating 72), Black (Slope 138, Rating 75)
-- Holes: 9 (front nine only, holes 1-9)
-- Generated: 2026-06-14
--
-- Corrections applied from source data:
--   - Hole 6: 'Bailouy' corrected to 'bailout'
--   - Hole 7: 'Gren' corrected to 'green' (reference_point)
--   - Hole 8 OB right: distance (0, 22) corrected to (22, null) per OB convention
--   - Hole 9: bunker labels corrected — rows 2-4 say 'right' but side = 'left'
-- --------------------------------------------------------


-- --------------------------------------------------------
-- COURSE
-- --------------------------------------------------------

INSERT INTO courses (id, name, golf_api_id, created_at)
VALUES (1, 'St Michaels', NULL, CURRENT_TIMESTAMP);


-- --------------------------------------------------------
-- TEES
-- --------------------------------------------------------

INSERT INTO tees (id, course_id, name, slope_rating, course_rating)
VALUES
    (1, 1, 'white', 133, 72.0),
    (2, 1, 'black', 138, 75.0);


-- --------------------------------------------------------
-- COURSE HOLES
-- One row per hole, shared across tees
-- Green centre coordinates are tee-independent
-- --------------------------------------------------------

INSERT INTO course_holes (id, course_id, hole_number, green_centre_lat, green_centre_lng)
VALUES
    (1, 1, 1, -33.98786758, 151.2434068),
    (2, 1, 2, -33.99063532, 151.2435173),
    (3, 1, 3, -33.99050281, 151.2448426),
    (4, 1, 4, -33.98823302, 151.2426275),
    (5, 1, 5, -33.98697879, 151.2464049),
    (6, 1, 6, -33.98590665, 151.251454),
    (7, 1, 7, -33.98644462, 151.246324),
    (8, 1, 8, -33.98516828, 151.2503391),
    (9, 1, 9, -33.98540139, 151.2465457);


-- --------------------------------------------------------
-- TEE HOLES
-- Par, stroke index, distance, and tee coordinates per tee
-- tee_id 1 = white, tee_id 2 = black
-- --------------------------------------------------------

INSERT INTO tee_holes (
    id,
    course_hole_id,
    tee_id,
    par,
    stroke_index,
    distance,
    tee_centre_lat,
    tee_centre_lng
)
VALUES
    -- Hole 1
    (1,  1, 1, 4,  2, 378, -33.98487831, 151.2452043),
    (2,  1, 2, 4,  1, 399, -33.98458084, 151.2451351),

    -- Hole 2
    (3,  2, 1, 4,  4, 303, -33.98821152, 151.2420351),
    (4,  2, 2, 4,  4, 317, -33.98815815, 151.2419869),

    -- Hole 3
    (5,  3, 1, 3, 12, 166, -33.99090978, 151.2435391),
    (6,  3, 2, 3, 10, 170, -33.99105211, 151.2431904),

    -- Hole 4
    (7,  4, 1, 4, 16, 291, -33.99009668, 151.2447786),
    (8,  4, 2, 4,  8, 346, -33.99029016, 151.245468),

    -- Hole 5
    (9,  5, 1, 3, 10, 170, -33.9878884,  151.2451818),
    (10, 5, 2, 3,  6, 203, -33.98816862, 151.2447983),

    -- Hole 6
    (11, 6, 1, 5, 18, 433, -33.98684377, 151.2469347),
    (12, 6, 2, 5, 16, 501, -33.98737911, 151.2464746),

    -- Hole 7
    (13, 7, 1, 5,  8, 472, -33.98535534, 151.2512957),
    (14, 7, 2, 5, 14, 492, -33.98530587, 151.251685),

    -- Hole 8
    (15, 8, 1, 4,  6, 390, -33.98613131, 151.2463419),
    (16, 8, 2, 4, 12, 398, -33.9861937,  151.2461965),

    -- Hole 9
    (17, 9, 1, 4, 14, 287, -33.98493167, 151.249598),
    (18, 9, 2, 4, 18, 298, -33.98491996, 151.2497401);


-- --------------------------------------------------------
-- HOLE POINTS OF INTEREST
-- course_hole_id references course_holes.id (1-9 = holes 1-9)
-- specific_tee: null = applies to all tees
-- poi_type: lowercase with underscores
-- distance_start/end: metres from reference_point
-- distance_end: null = unbounded (OB lines, scrub to green, etc)
-- --------------------------------------------------------

INSERT INTO hole_points_of_interest (
    course_hole_id,
    specific_tee,
    poi_type,
    side,
    reference_point,
    distance_start,
    distance_end,
    label
)
VALUES

    -- --------------------------------------------------------
    -- HOLE 1 — Par 4, 378m/399m
    -- --------------------------------------------------------
    (1, NULL,    'out_of_bounds', 'right', 'tee',   0,    NULL, 'OB all down the right side'),
    (1, NULL,    'penalty_area',  'left',  'green',  150,  NULL, 'Hazard starts 150m from green down the left side'),
    (1, NULL,    'out_of_bounds', 'long',  'green',  12,   NULL, 'Out of bounds 12m from the back edge of the green'),

    -- --------------------------------------------------------
    -- HOLE 2 — Par 4, 303m/317m
    -- --------------------------------------------------------
    (2, NULL,    'carry',         NULL,    'green',  165,  NULL, 'Forced carry from the tee leaves 165 to the green'),
    (2, NULL,    'out_of_bounds', 'right', 'tee',    0,    NULL, 'OB all down the right side'),
    (2, NULL,    'dense_scrub',   'left',  'green',  103,  NULL, 'Heavy scrub down the left side from 103m from the green. Avoid'),
    (2, NULL,    'bunker',        'right', 'green',  22.5, 9.45, 'Greenside bunker on the right 22.5m from the centre of the green. 13m wide'),
    (2, NULL,    'bunker',        'long',  'green',  0,    13,   'Greenside bunker long 13m from middle of green'),
    (2, NULL,    'bailout',       'left',  'green',  0,    17,   'Bailout area on the left 17m short of centre'),

    -- --------------------------------------------------------
    -- HOLE 3 — Par 3, 166m/170m
    -- --------------------------------------------------------
    (3, NULL,    'out_of_bounds', 'right', 'tee',    0,    NULL, 'OB all down the right side'),
    (3, NULL,    'dense_scrub',   'long',  'green',  0,    NULL, 'Dense scrub right behind the green. All downhill'),
    (3, NULL,    'dense_scrub',   'left',  'tee',    0,    NULL, 'Dense scrub left. Widens nearer the green'),
    (3, NULL,    'bailout',       'left',  'green',  24,   NULL, 'Safe landing 24m radius of the green. Prefer left'),

    -- --------------------------------------------------------
    -- HOLE 4 — Par 4, 291m/346m
    -- --------------------------------------------------------
    (4, 'black', 'carry',         NULL,    'tee',    0,    190,  'Minimum 190m blind carry to edge of fairway from black tees'),
    (4, NULL,    'water',         'right', 'green',  118,  63,   'Water down the right side. Starts 63m out from the green till 118m out'),
    (4, NULL,    'bunker',        'left',  'green',  117,  42,   'Bunkers all down the left side. Starts 42m out till 117m out from green'),
    (4, NULL,    'bunker',        NULL,    'green',  21,   12,   'Greenside bunker starts 21m out from green. 9m to cover'),
    (4, NULL,    'false_front',   NULL,    'green',  0,    18,   'Green has a false front 18m from centre. Will spin and roll off'),

    -- --------------------------------------------------------
    -- HOLE 5 — Par 3, 170m/203m
    -- --------------------------------------------------------
    (5, NULL,    'carry',         NULL,    'green',  47,   NULL, 'Sandy scrub carry ends 47m from green'),
    (5, NULL,    'bunker',        'left',  'green',  0,    13,   'Greenside bunker back left 13m from center'),
    (5, NULL,    'dense_scrub',   'left',  'green',  63,   16,   'Trees left 63m from green opening up 16m from green'),
    (5, NULL,    'bailout',       'right', 'green',  0,    16,   'Bailout area 16m from centre right'),
    (5, NULL,    'dense_scrub',   'long',  'green',  0,    21,   'Dense scrub behind the green 21m deep'),

    -- --------------------------------------------------------
    -- HOLE 6 — Par 5, 433m/501m
    -- --------------------------------------------------------
    (6, 'black', 'carry',         NULL,    'tee',    180,  NULL, '180m blind carry from black tees'),
    (6, 'white', 'carry',         NULL,    'tee',    120,  NULL, '120m carry from elevated white tee'),
    (6, NULL,    'water',         'right', 'green',  210,  52,   'Water starts 210m from green till 52m out'),
    (6, NULL,    'dense_scrub',   'right', 'green',  339,  227,  'Dense scrub right from 227m to 339m from the green'),
    (6, NULL,    'bunker',        'left',  'green',  173,  159,  'Fairway bunker 159-173m from green left'),
    (6, NULL,    'bunker',        'left',  'green',  205,  192,  'Fairway bunker 192-205m from green left'),
    (6, NULL,    'out_of_bounds', 'long',  'green',  25,   NULL, 'Out of bounds 25m from centre of green'),
    (6, NULL,    'bailout',       'right', 'green',  0,    21,   'Bailout area 21m short right of green'),

    -- --------------------------------------------------------
    -- HOLE 7 — Par 5, 472m/492m
    -- --------------------------------------------------------
    (7, NULL,    'penalty_area',  'right', 'green',  351,  289,  'Penalty area right from 351m to 289m from the green'),
    (7, NULL,    'bunker',        'right', 'green',  108,  76,   'Fairway bunkers right from 76m to 108m out from the green'),
    (7, NULL,    'bunker',        'left',  'green',  64,   53,   'Fairway bunkers left from 53m to 64m out from the green'),
    (7, NULL,    'bunker',        'right', 'green',  36,   22,   'Fairway bunkers right from 22m to 36m out from the green'),
    (7, NULL,    'dense_scrub',   'left',  'green',  97,   NULL, 'Dense scrub starting 97m out from the green on the left'),
    (7, NULL,    'dense_scrub',   'long',  'green',  29,   NULL, 'Dense scrub behind the green starting 29m from the middle of the green'),

    -- --------------------------------------------------------
    -- HOLE 8 — Par 4, 390m/398m
    -- --------------------------------------------------------
    (8, NULL,    'penalty_area',  'left',  'green',  180,  140,  'Penalty area left starting 180m from the green for 40m'),
    (8, NULL,    'penalty_area',  'right', 'green',  107,  47,   'Hazard on the right starting 47m from the green for 60m'),
    (8, NULL,    'bunker',        'right', 'green',  17,   9,    'Greenside bunker 17m from center on the right. 8m to carry'),
    -- NOTE: OB right inline with green — 22m gap from green centre to OB line.
    -- distance_start = 22 (where OB begins), distance_end = null (unbounded beyond that)
    (8, NULL,    'out_of_bounds', 'right', 'green',  22,   NULL, 'Out of bounds on the right inline with the green. 22m gap from green centre to OB'),

    -- --------------------------------------------------------
    -- HOLE 9 — Par 4, 287m/298m
    -- NOTE: Source data labels for bunker rows 2-4 said 'right' but side = 'left'.
    -- Labels corrected to match side values.
    -- --------------------------------------------------------
    (9, NULL,    'bunker',        'right', 'green',  92,   41,   'Fairway bunkers on the right starting 41m out from the green. 51m to carry'),
    (9, NULL,    'bunker',        'left',  'green',  82,   70,   'Fairway bunkers on the left starting 70m out. 12m to carry'),
    (9, NULL,    'bunker',        'left',  'green',  55,   47,   'Fairway bunkers on the left starting 47m out. 8m to carry'),
    (9, NULL,    'bunker',        'left',  'green',  38,   28,   'Fairway bunkers on the left starting 28m out. 10m to carry'),
    (9, NULL,    'dense_scrub',   'left',  'green',  38,   NULL, 'Dense scrub starting 38m out on the left'),
    (9, NULL,    'bailout',       'right', 'green',  10,   NULL, 'Bailout area on the right 10m from centre'),
    (9, NULL,    'false_front',   NULL,    'green',  0,    12,   'Green has a false front 12m from centre. Will roll off');