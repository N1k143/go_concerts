-- Locations
CREATE TABLE IF NOT EXISTS locations (
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

-- Concerts
CREATE TABLE IF NOT EXISTS concerts (
    id          BIGSERIAL PRIMARY KEY,
    artist      VARCHAR(50) NOT NULL,
    location_id BIGINT NOT NULL REFERENCES locations(id)
);

-- Shows
CREATE TABLE IF NOT EXISTS shows (
    id         BIGSERIAL PRIMARY KEY,
    concert_id BIGINT NOT NULL REFERENCES concerts(id),
    start      TIMESTAMPTZ NOT NULL,
    "end"      TIMESTAMPTZ NOT NULL
);

-- Reservations
CREATE TABLE IF NOT EXISTS reservations (
    id         BIGSERIAL PRIMARY KEY,
    token      VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Bookings
CREATE TABLE IF NOT EXISTS bookings (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    address    VARCHAR(100) NOT NULL,
    city       VARCHAR(100) NOT NULL,
    zip        VARCHAR(10) NOT NULL,
    country    VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tickets
CREATE TABLE IF NOT EXISTS tickets (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(10) NOT NULL,
    booking_id BIGINT NOT NULL REFERENCES bookings(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Location seat rows (tied to a show)
CREATE TABLE IF NOT EXISTS location_seat_rows (
    id      BIGSERIAL PRIMARY KEY,
    name    VARCHAR(50) NOT NULL,
    "order" SMALLINT NOT NULL,
    show_id BIGINT NOT NULL REFERENCES shows(id),
    UNIQUE (name, show_id),
    UNIQUE ("order", show_id)
);

-- Location seats
CREATE TABLE IF NOT EXISTS location_seats (
    id                   BIGSERIAL PRIMARY KEY,
    location_seat_row_id BIGINT NOT NULL REFERENCES location_seat_rows(id),
    number               SMALLINT NOT NULL,
    reservation_id       BIGINT REFERENCES reservations(id),
    ticket_id            BIGINT REFERENCES tickets(id),
    UNIQUE (location_seat_row_id, number)
);

-- ==================== SEED DATA ====================

INSERT INTO locations (id, name) VALUES
(1, 'Oper Graz'),
(2, 'Freilufthalle B'),
(3, 'Das Orpheum')
ON CONFLICT DO NOTHING;
SELECT setval('locations_id_seq', (SELECT MAX(id) FROM locations));

INSERT INTO concerts (id, artist, location_id) VALUES
(1, 'Opus', 1),
(2, 'Bilderbuch', 3),
(3, 'Wanda', 2),
(4, 'Christina Stürmer', 3)
ON CONFLICT DO NOTHING;
SELECT setval('concerts_id_seq', (SELECT MAX(id) FROM concerts));

INSERT INTO shows (id, concert_id, start, "end") VALUES
(1, 1, '2021-10-02 21:00:00+02', '2021-10-03 00:00:00+02'),
(2, 2, '2021-10-01 19:00:00+02', '2021-10-01 21:00:00+02'),
(3, 4, '2021-10-03 20:00:00+02', '2021-10-03 23:00:00+02'),
(4, 2, '2021-09-30 22:30:00+02', '2021-10-01 00:30:00+02'),
(5, 3, '2021-09-30 19:00:00+02', '2021-09-30 21:00:00+02'),
(6, 3, '2021-10-01 22:30:00+02', '2021-10-02 00:30:00+02')
ON CONFLICT DO NOTHING;
SELECT setval('shows_id_seq', (SELECT MAX(id) FROM shows));

-- Bookings (first 10 for seed)
INSERT INTO bookings (id, name, address, city, zip, country, created_at) VALUES
(1, 'Daniel Maierhofer', 'Astrid-Angerer-Weg 2c', 'Groß-Siegharts', '1070', 'Austria', '2021-08-09 13:41:25+02'),
(2, 'Ralph Sattler', 'Julian-Brunner-Gasse 55b', 'Korneuburg', '7000', 'Austria', '2021-06-27 18:03:26+02'),
(3, 'Mona Steinbauer', 'Schlagerstraße 369', 'Graz', '9170', 'Austria', '2021-01-06 09:09:00+01'),
(4, 'Ardith Schowalter', '6618 McLaughlin Ports', 'Kertzmannburgh', '63685', 'Oman', '2021-01-07 00:35:21+01'),
(5, 'Katheryn Funk', '8703 Ryleigh Prairie', 'North Reagan', '72830', 'Monaco', '2021-06-12 00:44:39+02'),
(6, 'Nina Führer', 'Steinbergergasse 55', 'Langenlois', '2700', 'Austria', '2021-05-21 03:29:06+02'),
(7, 'Kilian Nikolic', 'Martinring 9c', 'Haag', '8200', 'Austria', '2021-09-11 00:57:25+02'),
(8, 'Valentin Hartl', 'Joseph-Gangl-Gasse 707', 'Althofen', '1230', 'Austria', '2021-06-28 02:16:09+02'),
(9, 'Christian Baumgartner', 'Rathring 52', 'Seekirchen am Wallersee', '8605', 'Austria', '2021-04-20 03:58:15+02'),
(10, 'Raphael Fasching', 'Gratzerstraße 9a', 'Trofaiach', '4780', 'Austria', '2021-04-28 13:53:47+02')
ON CONFLICT DO NOTHING;
SELECT setval('bookings_id_seq', (SELECT MAX(id) FROM bookings));

-- Seed location_seat_rows for show 1 (Stalls + Terrace)
INSERT INTO location_seat_rows (id, name, "order", show_id) VALUES
(1, 'Stalls 01', 1, 1),(2, 'Stalls 02', 2, 1),(3, 'Stalls 03', 3, 1),
(4, 'Stalls 04', 4, 1),(5, 'Stalls 05', 5, 1),(6, 'Stalls 06', 6, 1),
(7, 'Stalls 07', 7, 1),(8, 'Stalls 08', 8, 1),(9, 'Stalls 09', 9, 1),
(10, 'Stalls 10', 10, 1),(11, 'Stalls 11', 11, 1),(12, 'Stalls 12', 12, 1),
(13, 'Stalls 13', 13, 1),(14, 'Stalls 14', 14, 1),(15, 'Stalls 15', 15, 1),
(16, 'Stalls 16', 16, 1),(17, 'Stalls 17', 17, 1),(18, 'Stalls 18', 18, 1),
(19, 'Terrace 1', 19, 1),(20, 'Terrace 2', 20, 1),(21, 'Terrace 3', 21, 1),
(22, 'Terrace 4', 22, 1),(23, 'Terrace 5', 23, 1),(24, 'Terrace 6', 24, 1),
-- show 5 rows
(25, '1', 1, 5),(26, '2', 2, 5),(27, '3', 3, 5),(28, '4', 4, 5),
(29, '5', 5, 5),(30, '6', 6, 5),(31, '7', 7, 5),(32, '8', 8, 5),
(33, '9', 9, 5),(34, '10', 10, 5),(35, '11', 11, 5),(36, '12', 12, 5),
(37, '13', 13, 5),
-- show 6 rows
(38, '1', 1, 6),(39, '2', 2, 6),(40, '3', 3, 6),(41, '4', 4, 6),
(42, '5', 5, 6),(43, '6', 6, 6),(44, '7', 7, 6),(45, '8', 8, 6),
(46, '9', 9, 6),(47, '10', 10, 6),(48, '11', 11, 6),(49, '12', 12, 6),
(50, '13', 13, 6),
-- show 2 rows (A-N)
(51, 'A', 1, 2),(52, 'B', 2, 2),(53, 'C', 3, 2),(54, 'D', 4, 2),
(55, 'E', 5, 2),(56, 'F', 6, 2),(57, 'G', 7, 2),(58, 'H', 8, 2),
(59, 'I', 9, 2),(60, 'J', 10, 2),(61, 'K', 11, 2),(62, 'L', 12, 2),
(63, 'M', 13, 2),(64, 'N', 14, 2),
-- show 3 rows (A-N)
(65, 'A', 1, 3),(66, 'B', 2, 3),(67, 'C', 3, 3),(68, 'D', 4, 3),
(69, 'E', 5, 3),(70, 'F', 6, 3),(71, 'G', 7, 3),(72, 'H', 8, 3),
(73, 'I', 9, 3),(74, 'J', 10, 3),(75, 'K', 11, 3),(76, 'L', 12, 3),
(77, 'M', 13, 3),(78, 'N', 14, 3),
-- show 4 rows (A-N)
(79, 'A', 1, 4),(80, 'B', 2, 4),(81, 'C', 3, 4),(82, 'D', 4, 4),
(83, 'E', 5, 4),(84, 'F', 6, 4),(85, 'G', 7, 4),(86, 'H', 8, 4),
(87, 'I', 9, 4),(88, 'J', 10, 4),(89, 'K', 11, 4),(90, 'L', 12, 4),
(91, 'M', 13, 4),(92, 'N', 14, 4)
ON CONFLICT DO NOTHING;
SELECT setval('location_seat_rows_id_seq', (SELECT MAX(id) FROM location_seat_rows));

-- Generate seats: show 1 rows 1-18 have 40 seats, rows 19-24 have 20 seats
-- show 5/6 rows 25-50 have 20 seats
-- show 2 rows 51-64 have 15-20 seats
-- show 3 rows 65-78 have 16 seats
-- show 4 rows 79-92 have 15 seats
DO $$
DECLARE
    r RECORD;
    i INT;
    max_seats INT;
BEGIN
    FOR r IN SELECT id, show_id, "order" FROM location_seat_rows ORDER BY id LOOP
        IF r.show_id = 1 THEN
            IF r."order" <= 18 THEN max_seats := 40;
            ELSE max_seats := 20; END IF;
        ELSIF r.show_id IN (5, 6) THEN
            max_seats := 20;
        ELSIF r.show_id = 2 THEN
            max_seats := 15;
        ELSIF r.show_id = 3 THEN
            max_seats := 16;
        ELSIF r.show_id = 4 THEN
            max_seats := 15;
        ELSE
            max_seats := 20;
        END IF;

        FOR i IN 1..max_seats LOOP
            INSERT INTO location_seats (location_seat_row_id, number)
            VALUES (r.id, i)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END LOOP;
END $$;

-- Add some pre-existing tickets for show 1 (to simulate booked seats)
-- booking 1 tickets: seats in row 1 (Stalls 01), numbers 1-4
DO $$
DECLARE
    t_id BIGINT;
    s_id BIGINT;
BEGIN
    -- booking 1: 4 tickets in show 1, row 1, seats 1-4
    FOR i IN 1..4 LOOP
        INSERT INTO tickets (code, booking_id) VALUES (
            upper(substring(md5(random()::text) for 10)),
            1
        ) RETURNING id INTO t_id;
        UPDATE location_seats SET ticket_id = t_id
        WHERE location_seat_row_id = 1 AND number = i;
    END LOOP;

    -- booking 2: 7 tickets in show 1, row 2, seats 1-7
    FOR i IN 1..7 LOOP
        INSERT INTO tickets (code, booking_id) VALUES (
            upper(substring(md5(random()::text) for 10)),
            2
        ) RETURNING id INTO t_id;
        UPDATE location_seats SET ticket_id = t_id
        WHERE location_seat_row_id = 2 AND number = i;
    END LOOP;
END $$;
