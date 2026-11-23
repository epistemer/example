-- Confab: the database of a conference organiser's back office.
-- Comments are the English labels models generates (table.*, enum.*).

CREATE TYPE event_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE talk_level AS ENUM ('beginner', 'intermediate', 'advanced');
CREATE TYPE talk_format AS ENUM ('talk', 'workshop', 'lightning', 'panel');
CREATE TYPE talk_state AS ENUM ('submitted', 'accepted', 'rejected', 'scheduled');
CREATE TYPE talk_topic AS ENUM ('go', 'web', 'data', 'security', 'operations', 'design');

CREATE TABLE staff (
    id         text        PRIMARY KEY,
    email      text        NOT NULL UNIQUE,
    name       text        NOT NULL,
    disabled   boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE staff IS 'Staff member';
COMMENT ON COLUMN staff.email IS 'Email address';

CREATE TABLE event (
    id              bigserial    PRIMARY KEY,
    name            text         NOT NULL,
    slug            text         NOT NULL UNIQUE,
    summary         text         NOT NULL DEFAULT '',
    status          event_status NOT NULL DEFAULT 'draft',
    starts_on       date         NOT NULL,
    ends_on         date         NOT NULL,
    online          boolean      NOT NULL DEFAULT false,
    capacity        integer      NOT NULL,
    venue_name      text,
    venue_street    text,
    venue_city      text,
    venue_country   text,
    venue_post_code text,
    latitude        double precision,
    longitude       double precision,
    currency        text         NOT NULL DEFAULT 'EUR',
    version         integer      NOT NULL DEFAULT 1,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now()
);
COMMENT ON TABLE event IS 'Event';
COMMENT ON COLUMN event.starts_on IS 'First day';
COMMENT ON COLUMN event.ends_on IS 'Last day';
COMMENT ON COLUMN event.online IS 'Online only';
COMMENT ON COLUMN event.venue_post_code IS 'Post code';

CREATE TABLE ticket_tier (
    id           bigserial     PRIMARY KEY,
    event_id     bigint        NOT NULL REFERENCES event ON DELETE CASCADE,
    position     integer       NOT NULL,
    name         text          NOT NULL,
    price        numeric(10,2) NOT NULL,
    quantity     integer       NOT NULL,
    sale_ends_on date
);
COMMENT ON TABLE ticket_tier IS 'Ticket tier';
COMMENT ON COLUMN ticket_tier.sale_ends_on IS 'On sale until';

CREATE TABLE speaker (
    id         bigserial   PRIMARY KEY,
    name       text        NOT NULL,
    email      text        NOT NULL UNIQUE,
    phone      text,
    country    text        NOT NULL,
    website    text,
    bio        text        NOT NULL DEFAULT '',
    photo_key  text,
    photo_name text,
    created_at timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE speaker IS 'Speaker';
COMMENT ON COLUMN speaker.bio IS 'Biography';
COMMENT ON COLUMN speaker.photo_key IS 'Photo';

CREATE TABLE talk (
    id               bigserial    PRIMARY KEY,
    event_id         bigint       NOT NULL REFERENCES event ON DELETE CASCADE,
    speaker_id       bigint       NOT NULL REFERENCES speaker,
    title            text         NOT NULL,
    abstract         text         NOT NULL,
    level            talk_level   NOT NULL,
    format           talk_format  NOT NULL,
    state            talk_state   NOT NULL DEFAULT 'submitted',
    topics           talk_topic[] NOT NULL DEFAULT '{}',
    starts_at        timestamptz,
    duration_minutes integer      NOT NULL,
    seats            integer,
    recorded         boolean,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now()
);
COMMENT ON TABLE talk IS 'Talk';
COMMENT ON COLUMN talk.event_id IS 'Event';
COMMENT ON COLUMN talk.speaker_id IS 'Speaker';
COMMENT ON COLUMN talk.starts_at IS 'Starts at';
COMMENT ON COLUMN talk.duration_minutes IS 'Duration (minutes)';
COMMENT ON COLUMN talk.seats IS 'Workshop seats';
COMMENT ON COLUMN talk.recorded IS 'Recorded';

CREATE TABLE talk_file (
    id           bigserial   PRIMARY KEY,
    talk_id      bigint      NOT NULL REFERENCES talk ON DELETE CASCADE,
    key          text        NOT NULL UNIQUE,
    name         text        NOT NULL,
    content_type text        NOT NULL,
    size         bigint      NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE talk_file IS 'Slides';

-- Views: models generates a listing method and a column set for each.

CREATE VIEW event_view AS
SELECT e.id, e.name, e.slug, e.status, e.starts_on, e.ends_on, e.online,
       e.capacity, e.venue_city, e.currency,
       (SELECT count(*) FROM talk t WHERE t.event_id = e.id)::integer AS talks,
       (SELECT min(price) FROM ticket_tier p WHERE p.event_id = e.id)::numeric(10,2) AS from_price,
       e.created_at
FROM event e;
COMMENT ON VIEW event_view IS 'Events';
COMMENT ON COLUMN event_view.from_price IS 'Tickets from';

CREATE VIEW speaker_view AS
SELECT s.id, s.name, s.email, s.country, s.website,
       (SELECT count(*) FROM talk t WHERE t.speaker_id = s.id)::integer AS talks,
       s.created_at
FROM speaker s;
COMMENT ON VIEW speaker_view IS 'Speakers';

CREATE VIEW talk_view AS
SELECT t.id, t.title, t.level, t.format, t.state, t.starts_at,
       t.duration_minutes, t.recorded,
       t.speaker_id, s.name AS speaker_name,
       t.event_id, e.name AS event_name,
       t.created_at
FROM talk t
JOIN speaker s ON s.id = t.speaker_id
JOIN event e ON e.id = t.event_id;
COMMENT ON VIEW talk_view IS 'Talks';
COMMENT ON COLUMN talk_view.speaker_id IS 'Speaker';
COMMENT ON COLUMN talk_view.speaker_name IS 'Speaker';
COMMENT ON COLUMN talk_view.event_id IS 'Event';
COMMENT ON COLUMN talk_view.event_name IS 'Event';
