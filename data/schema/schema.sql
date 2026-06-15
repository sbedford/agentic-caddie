DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS tees;
DROP TABLE IF EXISTS course_holes;
DROP TABLE IF EXISTS tee_holes;
DROP TABLE IF EXISTS hole_points_of_interest;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS player_clubs;
DROP TABLE IF EXISTS rounds;
DROP TABLE IF EXISTS holes;
DROP TABLE IF EXISTS shots;
DROP TABLE IF EXISTS commentary;

-- ============================================================
-- Golf Caddie — SQLite Schema
-- ============================================================
--
-- Design principles:
--   - Holes table is scorecard data only (flat, always populated)
--   - Shots table is the sequence layer (reconstructed for history,
--     manual for in-app entry, gps once tracking is live)
--   - Wind/temperature deferred — add per-hole once Stage 1 is working
--   - No free text entry anywhere — commentary is agent output only
--
-- Conventions:
--   - All distances in metres
--   - All enums are lowercase with underscores
--   - Nullable fields are intentionally nullable (not defaulting to false/0)
--
-- Deferred:
--   - wind_strength, wind_relative per hole (orientation-dependent)
--   - temperature per round
--   - GPS coordinates on shots (from_lat, from_lng, to_lat, to_lng)
--   - client_uuid, synced_at on shots (offline sync, Stage 3)
-- ============================================================


-- ============================================================
-- COURSES
-- ============================================================

CREATE TABLE courses (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    golf_api_id     TEXT UNIQUE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- One row per tee colour per course
-- slope_rating and course_rating are tee-specific, not course-level
CREATE TABLE tees (
    id              INTEGER PRIMARY KEY,
    course_id       INTEGER NOT NULL REFERENCES courses(id),
    name            TEXT NOT NULL,          -- 'white', 'black', 'red', 'gold'
    slope_rating    INTEGER,
    course_rating   REAL,

    UNIQUE(course_id, name)
);

-- Hole geography — independent of tee
-- Green coordinates are the same regardless of which tee is played
CREATE TABLE course_holes (
    id                  INTEGER PRIMARY KEY,
    course_id           INTEGER NOT NULL REFERENCES courses(id),
    hole_number         INTEGER NOT NULL,
    green_centre_lat    REAL,
    green_centre_lng    REAL,

    UNIQUE(course_id, hole_number)
);

-- Playing characteristics per tee per hole
-- Par and stroke index can differ by tee (rare but valid)
-- Tee coordinates recorded here as they are tee-specific
CREATE TABLE tee_holes (
    id                  INTEGER PRIMARY KEY,
    course_hole_id      INTEGER NOT NULL REFERENCES course_holes(id),
    tee_id              INTEGER NOT NULL REFERENCES tees(id),
    par                 INTEGER NOT NULL,
    stroke_index        INTEGER,
    distance            INTEGER NOT NULL,   -- metres from tee to green
    tee_centre_lat      REAL,
    tee_centre_lng      REAL,

    UNIQUE(course_hole_id, tee_id)
);

-- Points of interest per hole
-- Covers hazards, OB, carries, bailouts, elevation, false fronts
-- shared across tees by default, specific_tee overrides where needed
--
-- distance conventions:
--   reference_point = 'tee'   → distance measured from tee towards green
--   reference_point = 'green' → distance measured from green centre outward
--   distance_end = null       → unbounded (OB lines, scrub running to green)
--   side = 'long'             → beyond the green
--
-- poi_type values:
--   'out_of_bounds'  → OB lines, always unbounded (distance_end = null)
--   'penalty_area'   → water, ditches, red/yellow stake areas
--   'bunker'         → fairway or greenside bunkers
--   'dense_scrub'    → trees, gorse, scrub — high penalty risk
--   'carry'          → forced carry distance
--   'bailout'        → safe miss zone
--   'elevation'      → significant elevation change affecting club selection
--   'false_front'    → green feature causing roll-off
CREATE TABLE hole_points_of_interest (
    id                  INTEGER PRIMARY KEY,
    course_hole_id      INTEGER NOT NULL REFERENCES course_holes(id),
    specific_tee        TEXT,               -- null = all tees
                                            -- 'white', 'black' = tee-specific only
    poi_type            TEXT NOT NULL,
    side                TEXT,               -- 'left', 'right', 'long', 'front', 'both'
                                            -- null for carry, elevation, false_front
    reference_point     TEXT,               -- 'tee' or 'green'
    distance_start      REAL,               -- metres from reference point
    distance_end        REAL,               -- null = unbounded
    label               TEXT NOT NULL       -- plain english, primary agent-facing content
);


-- ============================================================
-- PLAYER
-- ============================================================

CREATE TABLE players (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    handicap        REAL,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Club inventory with date ranges
-- added_date/removed_date scope distance calculations to the correct equipment era
-- A new club gets a new row — old row gets removed_date set
CREATE TABLE player_clubs (
    id                  INTEGER PRIMARY KEY,
    player_id           INTEGER NOT NULL REFERENCES players(id),
    club_name           TEXT NOT NULL,      -- 'driver', '3_hybrid', '3i', '4i', '5i',
                                            --  '6i', '7i', '8i', '9i', 'pw', 'gw', 'sw'
    added_date          DATE NOT NULL,
    removed_date        DATE,               -- null if still in bag

    -- distance model
    -- manually seeded to start (sample_size = 0)
    -- progressively replaced by GPS-derived values as rounds accumulate
    carry_avg           REAL,               -- mean carry, null until GPS data available
    carry_reliable      REAL,               -- 25th percentile — the planning number
    carry_max           REAL,               -- 90th percentile, null until GPS data available
    dispersion_avg_m    REAL,               -- average lateral miss in metres
    dispersion_bias     TEXT,               -- 'left', 'right', 'straight'
    sample_size         INTEGER NOT NULL DEFAULT 0,
                                            -- 0 = manually seeded
                                            -- grows with GPS-tracked shots
    calculated_at       TIMESTAMP,

    UNIQUE(player_id, club_name, added_date)
);


-- ============================================================
-- ROUNDS
-- ============================================================

-- round_type values: 'competition', 'social', 'practice'
-- tees references tees.name for the course being played
-- total_score/points/putts are denormalised for query performance
--   — derived from holes but stored to avoid aggregation on every read
--
-- deferred: wind_strength, wind_direction, temperature
--   add per-hole once Stage 1 is working (wind is orientation-dependent,
--   round-level wind direction is not meaningful without hole bearing)
CREATE TABLE rounds (
    id              INTEGER PRIMARY KEY,
    player_id       INTEGER NOT NULL REFERENCES players(id),
    course_id       INTEGER NOT NULL REFERENCES courses(id),
    played_at       DATE NOT NULL,
    tees            TEXT NOT NULL,          -- 'white', 'black'
    round_type      TEXT NOT NULL,          -- 'competition', 'social', 'practice'
    total_score     INTEGER,
    total_points    INTEGER,
    total_putts     INTEGER,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- ============================================================
-- HOLES
-- Scorecard data only — flat, always populated
-- Shot narrative lives in the shots table
-- ============================================================

-- flag_position values:
--   'front_left', 'front_centre', 'front_right'
--   'middle_left', 'middle_centre', 'middle_right'
--   'back_left', 'back_centre', 'back_right'
--
-- gir: true = green hit in regulation
--      false + scramble_attempt = true → scramble situation
--      false + scramble_attempt = false → didn't attempt green in regulation
--        (punch out, layup forced by position)
--
-- penalty: any penalty stroke on the hole, regardless of where it occurred
CREATE TABLE holes (
    id                  INTEGER PRIMARY KEY,
    round_id            INTEGER NOT NULL REFERENCES rounds(id),
    course_hole_id      INTEGER NOT NULL REFERENCES course_holes(id),
    hole_number         INTEGER NOT NULL,
    flag_position       TEXT,
    score               INTEGER,
    points              INTEGER,
    putts               INTEGER,
    gir                 BOOLEAN,
    scramble_save       BOOLEAN,
    penalty             BOOLEAN,

    UNIQUE(round_id, hole_number)
);


-- ============================================================
-- SHOTS
-- One row per shot, sequential within each hole
-- ============================================================

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
--
-- source values:
--   'reconstructed' → derived from spreadsheet hole totals
--                     (tee/approach ends are reliable, middle shots are unknown)
--   'manual'        → entered shot by shot in the app
--   'gps'           → GPS tracked (Stage 3)
--
-- GPS columns deferred — added when Stage 3 tracking is implemented:
--   from_lat, from_lng, to_lat, to_lng, carry_distance
--   client_uuid (device-generated dedup key for offline sync)
--   recorded_at, synced_at
CREATE TABLE shots (
    id              INTEGER PRIMARY KEY,
    hole_id         INTEGER NOT NULL REFERENCES holes(id),
    shot_number     INTEGER NOT NULL,
    shot_type       TEXT NOT NULL,
    club            TEXT,
    result          TEXT,
    miss            TEXT,
    strike_quality  TEXT,
    source          TEXT NOT NULL DEFAULT 'manual',

    UNIQUE(hole_id, shot_number)
);


-- ============================================================
-- AGENT COMMENTARY
-- Generated by agent post-hole and post-round
-- Never player-entered
-- ============================================================

-- scope values: 'hole', 'round'
-- scope_id: hole_id or round_id depending on scope
CREATE TABLE commentary (
    id              INTEGER PRIMARY KEY,
    scope           TEXT NOT NULL,
    scope_id        INTEGER NOT NULL,
    content         TEXT NOT NULL,
    generated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- ============================================================
-- INDEXES
-- Added for the most common agent query patterns
-- ============================================================

-- Round history queries (most common agent entry point)
CREATE INDEX IF  NOT EXISTS idx_rounds_player_date
    ON rounds(player_id, played_at DESC);

-- Hole lookup within a round
CREATE INDEX IF  NOT EXISTS  idx_holes_round
    ON holes(round_id, hole_number);

-- Shot sequence lookup (always accessed via hole)
CREATE INDEX IF  NOT EXISTS  idx_shots_hole
    ON shots(hole_id, shot_number);

-- POI lookup by hole (agent fetches these for every caddie call)
CREATE INDEX IF  NOT EXISTS  idx_poi_hole
    ON hole_points_of_interest(course_hole_id);

-- Commentary lookup by scope
CREATE INDEX IF  NOT EXISTS  idx_commentary_scope
    ON commentary(scope, scope_id);