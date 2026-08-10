import {
  buildDownsamplingSql,
  buildGapfillSql,
  buildLatestSql,
  buildRawSql,
  buildSql,
  buildWindowSql,
} from '../src/sqlBuilder';
import type { KwdbQuery } from '../src/types';

const base: KwdbQuery = { refId: 'A', mode: 'downsampling', format: 'time_series' };

describe('sqlBuilder', () => {
  it('builds downsampling SQL with a qualified table, tags and one metric', () => {
    const query: KwdbQuery = {
      ...base,
      table: 'ts_db.sensor_data',
      timeColumn: 'ts',
      interval: '5m',
      tags: ['device_id'],
      metrics: [{ column: 'temperature', aggregation: 'avg' }],
    };

    expect(buildSql(query)).toBe(
      [
        'SELECT time_bucket("ts", \'5m\') AS "time",',
        '       "device_id",',
        '       avg("temperature") AS "avg_temperature"',
        'FROM "ts_db"."sensor_data"',
        'WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo',
        'GROUP BY "time", "device_id"',
        'ORDER BY "time", "device_id"',
      ].join('\n')
    );
  });

  it('uses the default 5m interval and handles multiple metrics, custom alias and none aggregation', () => {
    const query: KwdbQuery = {
      ...base,
      table: 'sensors',
      timeColumn: 'k_timestamp',
      tags: [],
      metrics: [
        { column: 'humidity', aggregation: 'none' },
        { column: 'temperature', aggregation: 'max', alias: 'peak' },
      ],
    };

    expect(buildDownsamplingSql(query)).toBe(
      [
        'SELECT time_bucket("k_timestamp", \'5m\') AS "time",',
        '       "humidity",',
        '       max("temperature") AS "peak"',
        'FROM "sensors"',
        'WHERE "k_timestamp" >= $__timeFrom AND "k_timestamp" <= $__timeTo',
        'GROUP BY "time"',
        'ORDER BY "time"',
      ].join('\n')
    );
  });

  it('quotes identifiers containing double quotes', () => {
    const query: KwdbQuery = {
      ...base,
      table: 'odd"table',
      timeColumn: 'we"ird',
      tags: ['tag"x'],
      metrics: [{ column: 'val"ue', aggregation: 'sum' }],
    };

    expect(buildDownsamplingSql(query)).toContain('FROM "odd""table"');
    expect(buildDownsamplingSql(query)).toContain('"we""ird"');
    expect(buildDownsamplingSql(query)).toContain('"tag""x"');
    expect(buildDownsamplingSql(query)).toContain('sum("val""ue")');
  });

  it('builds gapfill SQL with linear interpolation by default', () => {
    const query: KwdbQuery = {
      ...base,
      mode: 'gapfill',
      table: 'ts_db.sensor_data',
      timeColumn: 'ts',
      interval: '5m',
      tags: ['device_id'],
      metrics: [{ column: 'temperature', aggregation: 'avg' }],
    };

    expect(buildGapfillSql(query)).toBe(
      [
        'SELECT time_bucket_gapfill("ts", \'5m\') AS "time",',
        '       "device_id",',
        '       interpolate(avg("temperature"), \'linear\') AS "avg_temperature"',
        'FROM "ts_db"."sensor_data"',
        'WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo',
        'GROUP BY "time", "device_id"',
        'ORDER BY "time", "device_id"',
      ].join('\n')
    );
  });

  it.each(['PREV', 'NEXT', 'NULL'] as const)('builds gapfill SQL with %s interpolation', (mode) => {
    const query: KwdbQuery = {
      ...base,
      mode: 'gapfill',
      table: 'sensors',
      timeColumn: 'ts',
      interval: '15m',
      interpolateMode: mode,
      metrics: [{ column: 'pressure', aggregation: 'avg' }],
    };

    expect(buildGapfillSql(query)).toContain(`interpolate(avg("pressure"), ${mode}) AS "avg_pressure"`);
  });

  it('builds latest values SQL with last() and groups by tags', () => {
    const query: KwdbQuery = {
      ...base,
      mode: 'latest',
      table: 'ts_db.sensor_data',
      timeColumn: 'ts',
      latestFunc: 'last',
      tags: ['device_id'],
      metrics: [{ column: 'temperature', aggregation: 'none' }],
    };

    expect(buildLatestSql(query)).toBe(
      [
        'SELECT "device_id",',
        '       last("temperature") AS "latest_temperature",',
        '       last("ts") AS "time"',
        'FROM "ts_db"."sensor_data"',
        'WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo',
        'GROUP BY "device_id"',
      ].join('\n')
    );
  });

  it('builds latest values SQL with first_row() and no GROUP BY when tags are empty', () => {
    const query: KwdbQuery = {
      ...base,
      mode: 'latest',
      table: 'sensors',
      timeColumn: 'ts',
      latestFunc: 'first_row',
      metrics: [{ column: 'temperature', aggregation: 'none', alias: 'oldest_temp' }],
    };

    expect(buildLatestSql(query)).toBe(
      [
        'SELECT first_row("temperature") AS "oldest_temp",',
        '       last("ts") AS "time"',
        'FROM "sensors"',
        'WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo',
      ].join('\n')
    );
  });

  it('builds TIME_WINDOW SQL with slide interval', () => {
    const query: KwdbQuery = {
      ...base,
      mode: 'window',
      table: 'ts_db.sensor_data',
      timeColumn: 'ts',
      windowType: 'TIME_WINDOW',
      windowInterval: '1h',
      windowSlide: '15m',
      tags: ['device_id'],
      metrics: [{ column: 'temperature', aggregation: 'avg' }],
    };

    expect(buildWindowSql(query)).toBe(
      [
        'SELECT first("ts") AS "time",',
        '       "device_id",',
        '       avg("temperature") AS "avg_temperature"',
        'FROM "ts_db"."sensor_data"',
        'WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo',
        'GROUP BY "device_id", TIME_WINDOW("ts", \'1h\', \'15m\')',
        'ORDER BY "time", "device_id"',
      ].join('\n')
    );
  });

  it('builds SESSION_WINDOW, EVENT_WINDOW, COUNT_WINDOW and STATE_WINDOW SQL', () => {
    const session: KwdbQuery = {
      ...base,
      mode: 'window',
      table: 'sensors',
      timeColumn: 'ts',
      windowType: 'SESSION_WINDOW',
      windowInterval: '5m',
      metrics: [{ column: 'speed', aggregation: 'avg' }],
    };
    expect(buildWindowSql(session)).toContain('first("ts") AS "time"');
    expect(buildWindowSql(session)).toContain('GROUP BY SESSION_WINDOW("ts", \'5m\')');

    const event: KwdbQuery = {
      ...session,
      windowType: 'EVENT_WINDOW',
      eventStartCond: 'temp > 100',
      eventEndCond: 'temp <= 100',
    };
    expect(buildWindowSql(event)).toContain('GROUP BY EVENT_WINDOW(temp > 100, temp <= 100)');

    const count: KwdbQuery = {
      ...session,
      windowType: 'COUNT_WINDOW',
      windowInterval: '100',
      windowSlide: '10',
    };
    expect(buildWindowSql(count)).toContain('GROUP BY COUNT_WINDOW(100, 10)');

    const state: KwdbQuery = {
      ...session,
      windowType: 'STATE_WINDOW',
      eventStartCond: 'CASE WHEN voltage >= 225 THEN 1 ELSE 0 END',
    };
    expect(buildWindowSql(state)).toContain(
      'GROUP BY STATE_WINDOW(CASE WHEN voltage >= 225 THEN 1 ELSE 0 END)'
    );
  });

  it('passes raw SQL through unchanged', () => {
    const query: KwdbQuery = {
      ...base,
      mode: 'raw',
      rawSql: 'SELECT "ts", "temperature" FROM "sensors" WHERE "ts" >= $__timeFrom',
    };

    expect(buildRawSql(query)).toBe(query.rawSql);
    expect(buildSql(query)).toBe(query.rawSql);
  });

  it('returns an empty string when required visual-builder fields are missing', () => {
    expect(buildSql({ ...base, table: 'sensors' })).toBe('');
    expect(buildSql({ ...base, timeColumn: 'ts', metrics: [{ column: 'temperature', aggregation: 'avg' }] })).toBe('');
    expect(buildSql({ ...base, table: 'sensors', timeColumn: 'ts' })).toBe('');
    expect(buildSql({ ...base, table: 'sensors', timeColumn: 'ts', metrics: [{ column: '', aggregation: 'avg' }] })).toBe(
      ''
    );
    expect(
      buildWindowSql({
        ...base,
        mode: 'window',
        table: 'sensors',
        timeColumn: 'ts',
        windowType: 'EVENT_WINDOW',
        metrics: [{ column: 'temperature', aggregation: 'avg' }],
      })
    ).toBe('');
  });
});
