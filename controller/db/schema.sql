CREATE TABLE IF NOT EXISTS zone_list (
    id SERIAL NOT NULL,
    zone_label VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS event_schedule (
    id SERIAL NOT NULL,
    start_time TIME NOT NULL,
    duration_minutes INT NOT NULL,
    day_of_week INT NOT NULL,
    event_members VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);