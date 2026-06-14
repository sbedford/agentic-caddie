-- --------------------------------------------------------
-- COURSES
-- --------------------------------------------------------

CREATE TABLE IF NOT EXISTS courses (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    golf_api_id     TEXT UNIQUE,
    Created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tees (
    id              INTEGER PRIMARY KEY,
    course_id       INTEGER NOT NULL REFERENCES courses(id),
    name            TEXT NOT NULL,      -- 'white', 'black', 'red', 'gold'
    slope_rating    INTEGER,
    course_rating   REAL,

    UNIQUE(course_id, name)
);

CREATE TABLE IF NOT EXISTS course_holes (
    id                  INTEGER PRIMARY KEY,
    course_id           INTEGER NOT NULL REFERENCES courses(id),
    hole_number         INTEGER NOT NULL,
    green_centre_lat    REAL,
    green_centre_lng    REAL,

    UNIQUE(course_id, hole_number)
);

CREATE TABLE IF NOT EXISTS tee_holes (
    id                  INTEGER PRIMARY KEY,
    course_hole_id      INTEGER NOT NULL REFERENCES course_holes(id),
    tee_id              INTEGER NOT NULL REFERENCES tees(id),
    par                 INTEGER NOT NULL,
    stroke_index        INTEGER,
    distance            INTEGER NOT NULL,   -- metres from this tee to green
    tee_centre_lat    REAL,
    tee_centre_lng    REAL,

    UNIQUE(course_hole_id, tee_id)
);

CREATE TABLE IF NOT EXISTS hole_points_of_interest (
    id                  INTEGER PRIMARY KEY,
    course_hole_id      INTEGER NOT NULL REFERENCES course_holes(id),
    specific_tee        TEXT,               -- null = all tees, 'white'/'black' = tee-specific
    poi_type            TEXT NOT NULL,      -- 'out_of_bounds', 'penalty_area', 'bunker',
                                            --  'dense_scrub', 'carry', 'bailout',
                                            --  'elevation', 'false_front'
    side                TEXT,               -- 'left', 'right', 'long', 'front', 'both'
    reference_point     TEXT,               -- 'tee', 'green'
    distance_start      REAL,
    distance_end        REAL,               -- null = unbounded
    label               TEXT NOT NULL
);


-- --------------------------------------------------------
-- PLAYER
-- --------------------------------------------------------


CREATE TABLE IF NOT EXISTS players (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    handicap        REAL,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS player_clubs (
    id              INTEGER PRIMARY KEY,
    player_id       INTEGER NOT NULL REFERENCES players(id),
    club_name       TEXT NOT NULL,      -- 'driver', '3_hybrid', '3i', '4i', '5i',
                                        --  '6i', '7i', '8i', '9i', 'pw', 'gw', 'sw'
 -- carry distances in metres
    carry_avg           REAL,           -- mean, null until GPS data available
    carry_reliable      REAL,           -- 25th percentile, seed manually to start
    carry_max           REAL,           -- 90th percentile, null until GPS data available

    -- dispersion
    dispersion_avg_m    REAL,           -- average lateral miss in metres
    dispersion_bias     TEXT,           -- 'left', 'right', 'straight'

    -- confidence
    sample_size         INTEGER NOT NULL DEFAULT 0,
                                        -- 0 = manually seeded, grows with GPS data
    calculated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(player_id, club_name)
);

-- --------------------------------------------------------
-- ROUNDS
-- --------------------------------------------------------

CREATE TABLE IF NOT EXISTS rounds (
    id              INTEGER PRIMARY KEY,
    player_id       INTEGER NOT NULL REFERENCES players(id),
    course_id       INTEGER NOT NULL REFERENCES courses(id),
    played_at       DATE NOT NULL,
    tees            TEXT NOT NULL,  -- 'white', 'black'
    round_type      TEXT NOT NULL,  -- 'competition', 'social', 'practice'
    wind_strength   TEXT,           -- 'none', 'light', 'moderate', 'strong'
    wind_direction  TEXT,           -- 'into', 'downwind', 'cross_left',
                                    --  'cross_right', 'variable'
    temperature     TEXT,           -- 'cold', 'mild', 'warm'

    -- denormalised summaries for query performance
    total_score     INTEGER,
    total_points    INTEGER,
    total_putts     INTEGER,

    Created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- --------------------------------------------------------
-- HOLES
-- --------------------------------------------------------

CREATE TABLE IF NOT EXISTS holes (
    id               INTEGER PRIMARY KEY,
    round_id         INTEGER NOT NULL REFERENCES rounds(id),
    course_hole_id   INTEGER NOT NULL REFERENCES course_holes(id),
    hole_number      INTEGER NOT NULL,
    flag_position    TEXT,           -- 'front_left', 'front_centre', 'front_right',
                                     --  'middle_left', 'middle_centre', 'middle_right',
                                     --  'back_left', 'back_centre', 'back_right'
    score            INTEGER,
    points           INTEGER,
    putts            INTEGER,
    gir              BOOLEAN,
    fairway_hit      BOOLEAN,        -- null for par 3s
    scramble_attempt BOOLEAN,
    scramble_save    BOOLEAN,
    penalty          BOOLEAN,

    UNIQUE(round_id, hole_number)
);


-- --------------------------------------------------------
-- SHOTS
-- --------------------------------------------------------

CREATE TABLE IF NOT EXISTS shots (
    id              INTEGER PRIMARY KEY,
    client_uuid     TEXT UNIQUE,        -- device-generated, prevents duplicate sync
    hole_id         INTEGER NOT NULL REFERENCES holes(id),
    shot_number     INTEGER NOT NULL,
    shot_type       TEXT NOT NULL,      -- 'tee', 'layup', 'approach',
                                        --  'chip', 'bunker', 'putt'
    club            TEXT,               -- null for putts if not tracking

    -- GPS positions
    from_lat        REAL NOT NULL,
    from_lng        REAL NOT NULL,
    to_lat          REAL,               -- null until player walks to ball
    to_lng          REAL,
    carry_distance  REAL,               -- computed server-side from coordinates

    -- outcome
    result          TEXT,               -- 'fairway', 'rough', 'bunker',
                                        --  'green', 'ob', 'hazard', 'hole'
    strike_quality  TEXT,               -- 'clean', 'thin', 'fat', 'shank'

    recorded_at     TIMESTAMP NOT NULL,
    synced_at       TIMESTAMP,          -- null until synced to server

    UNIQUE(hole_id, shot_number)
);


-- --------------------------------------------------------
-- AGENT COMMENTARY
-- Generated by agent post-hole and post-round
-- Never player-entered
-- --------------------------------------------------------

CREATE TABLE IF NOT EXISTS commentary (
    id              INTEGER PRIMARY KEY,
    scope           TEXT NOT NULL,      -- 'hole' or 'round'
    scope_id        INTEGER NOT NULL,   -- hole_id or round_id
    content         TEXT NOT NULL,      -- agent-generated text
    generated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);