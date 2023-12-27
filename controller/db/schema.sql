CREATE TABLE IF NOT EXISTS zone_list (
    id SERIAL NOT NULL,
    zone_label VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS weekdays (
    id SERIAL NOT NULL,
    day_label VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)    
);

INSERT INTO weekdays (day_label) VALUES ('Sunday');
INSERT INTO weekdays (day_label) VALUES ('Monday');
INSERT INTO weekdays (day_label) VALUES ('Tuesday');
INSERT INTO weekdays (day_label) VALUES ('Wednesday');
INSERT INTO weekdays (day_label) VALUES ('Thursday');
INSERT INTO weekdays (day_label) VALUES ('Friday');
INSERT INTO weekdays (day_label) VALUES ('Saturday');

CREATE TABLE IF NOT EXISTS event_schedule (
    id SERIAL NOT NULL,
    zone_id INT NOT NULL,
    start_time TIME NOT NULL,
    duration_minutes INT NOT NULL,
    day_of_week INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (zone_id) REFERENCES zone_list(id),
    FOREIGN KEY (day_of_week) REFERENCES weekdays(id)
);