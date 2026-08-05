-- KWDB demo time-series database and sensor sample data for the Grafana data source.
CREATE TS DATABASE IF NOT EXISTS demo_ts;

CREATE TABLE IF NOT EXISTS demo_ts.sensors (
    ts TIMESTAMP NOT NULL,
    temperature DOUBLE,
    humidity DOUBLE,
    voltage DOUBLE
) TAGS (
    device_id INT NOT NULL,
    location VARCHAR(100)
) PRIMARY TAGS (device_id);

INSERT INTO demo_ts.sensors (ts, temperature, humidity, voltage, device_id, location) VALUES
    (NOW() - INTERVAL '15 minutes', 21.4, 58.1, 219.8, 1, 'room-a'),
    (NOW() - INTERVAL '15 minutes', 22.0, 57.6, 220.2, 2, 'room-b'),
    (NOW() - INTERVAL '10 minutes', 21.8, 57.9, 220.0, 1, 'room-a'),
    (NOW() - INTERVAL '10 minutes', 22.4, 57.2, 220.4, 2, 'room-b'),
    (NOW() - INTERVAL '5 minutes',  22.1, 57.4, 220.1, 1, 'room-a'),
    (NOW() - INTERVAL '5 minutes',  22.7, 56.9, 220.5, 2, 'room-b'),
    (NOW() - INTERVAL '2 minutes',  22.3, 57.0, 220.2, 1, 'room-a'),
    (NOW() - INTERVAL '2 minutes',  22.9, 56.6, 220.6, 2, 'room-b'),
    (NOW() - INTERVAL '1 minute',   22.5, 56.8, 220.3, 1, 'room-a'),
    (NOW() - INTERVAL '1 minute',   23.1, 56.4, 220.7, 2, 'room-b'),
    (NOW() - INTERVAL '30 seconds', 22.6, 56.5, 220.3, 3, 'room-c'),
    (NOW() - INTERVAL '15 seconds', 22.7, 56.3, 220.4, 3, 'room-c');
