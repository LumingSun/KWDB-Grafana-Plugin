import type { ColumnInfo } from '../../../types';
import { AGGREGATIONS, getAllowedAggregations, getAggregationsForColumn } from './aggregations';

function makeColumn(overrides: Partial<ColumnInfo> = {}): ColumnInfo {
  return {
    name: 'voltage',
    type: 'FLOAT8',
    isTag: false,
    isTimeColumn: false,
    isPrimaryTag: false,
    ...overrides,
  };
}

describe('getAllowedAggregations', () => {
  it('shows every aggregation for numeric columns', () => {
    for (const type of ['FLOAT8', 'INT4', 'BIGINT', 'DOUBLE', 'DECIMAL(18,2)', 'SMALLINT', 'REAL']) {
      expect(getAllowedAggregations(makeColumn({ type }))).toEqual(AGGREGATIONS);
    }
  });

  it('only shows count and none for string columns', () => {
    for (const type of ['VARCHAR(32)', 'CHAR(10)', 'TEXT', 'STRING']) {
      expect(getAllowedAggregations(makeColumn({ name: 'location', type }))).toEqual(['count', 'none']);
    }
  });

  it('treats tag columns as strings even with a numeric declared type', () => {
    expect(getAllowedAggregations(makeColumn({ name: 'device_id', type: 'INT4', isTag: true }))).toEqual([
      'count',
      'none',
    ]);
  });

  it('shows count, min, max and none for time columns', () => {
    for (const type of ['TIMESTAMPTZ(3)', 'TIMESTAMP', 'K_TIMESTAMP', 'TIME(3)']) {
      expect(getAllowedAggregations(makeColumn({ name: 'time', type, isTimeColumn: true }))).toEqual([
        'count',
        'min',
        'max',
        'none',
      ]);
    }
  });

  it('falls back to every aggregation for unknown columns or types', () => {
    expect(getAllowedAggregations(undefined)).toEqual(AGGREGATIONS);
    expect(getAllowedAggregations(makeColumn({ type: 'SOMETHING_ELSE' }))).toEqual(AGGREGATIONS);
  });
});

describe('getAggregationsForColumn', () => {
  const columns = [
    makeColumn({ name: 'voltage', type: 'FLOAT8' }),
    makeColumn({ name: 'device_id', type: 'VARCHAR(64)', isTag: true }),
  ];

  it('filters by the named column', () => {
    expect(getAggregationsForColumn(columns, 'voltage')).toEqual(AGGREGATIONS);
    expect(getAggregationsForColumn(columns, 'device_id')).toEqual(['count', 'none']);
  });

  it('returns every aggregation before a column is chosen', () => {
    expect(getAggregationsForColumn(columns)).toEqual(AGGREGATIONS);
    expect(getAggregationsForColumn(columns, '')).toEqual(AGGREGATIONS);
  });
});
