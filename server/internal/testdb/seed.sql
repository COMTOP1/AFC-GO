INSERT INTO teams (name, description, league, division, active, youth, ages) VALUES
    ('First Team', 'Senior men', 'Hellenic League', 'Division One East', TRUE, FALSE, 99),
    ('Under 12s', 'Youth squad', 'Youth League', 'U12', TRUE, TRUE, 12),
    ('Old Team', NULL, NULL, NULL, FALSE, FALSE, 99);

INSERT INTO users (name, email, phone, team_id, role, file_name) VALUES
    ('Web Master', 'webmaster@example.test', '01234 567890', 0, 'WEBMASTER', NULL),
    ('Team Manager', 'manager@example.test', NULL, 1, 'MANAGER', NULL),
    ('Club Secretary', 'secretary@example.test', NULL, 0, 'CLUB_SECRETARY', 'user/secretary.png');

INSERT INTO players (name, file_name, date_of_birth, position, captain, team_id) VALUES
    ('Adult Player', 'player/adult.png', '1990-05-01', 'Striker', TRUE, 1),
    ('Youth Player', 'player/youth.png', '2014-03-02', 'Keeper', FALSE, 2),
    ('Young Senior', 'player/young.png', CURRENT_DATE - INTERVAL '16 years', 'Defender', FALSE, 1);

INSERT INTO sponsors (name, website, file_name, purpose, team_id) VALUES
    ('Club Sponsor', 'https://sponsor.example.test', 'sponsor/club.png', 'Kit', 'A'),
    ('Team Sponsor', NULL, 'sponsor/team.png', NULL, '1');

INSERT INTO affiliations (name, website, file_name) VALUES
    ('County FA', 'https://fa.example.test', 'affiliation/fa.png');

INSERT INTO documents (name, file_name) VALUES ('Club Rules', 'document/rules.pdf');

INSERT INTO images (file_name, caption) VALUES ('gallery/one.jpg', 'Match day');

INSERT INTO news (title, file_name, content, date) VALUES
    ('Season opener', 'news/opener.jpg', '<p>We won!</p>', NOW() - INTERVAL '2 days'),
    ('Training update', NULL, '<p>Tuesday 7pm</p>', NOW() - INTERVAL '1 day');

INSERT INTO whatson (title, file_name, content, date, date_of_event) VALUES
    ('Presentation night', NULL, '<p>Clubhouse</p>', NOW(), CURRENT_DATE + 30),
    ('Summer BBQ', NULL, '<p>Done</p>', NOW(), CURRENT_DATE - 30);

INSERT INTO programme_seasons (season) VALUES ('2025-26');

INSERT INTO programmes (name, file_name, date_of_programme, programme_season_id) VALUES
    ('Opening day programme', 'programme/opening.pdf', CURRENT_DATE - 10, 1);

INSERT INTO settings (id, setting_text) VALUES
    ('displayEmail', 'hello@example.test'),
    ('infoContent', '<div>Club history</div>'),
    ('visitorCount', '42');
