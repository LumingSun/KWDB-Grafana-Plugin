import { buildVariableQueryString, parseVariableQuery } from '../src/variableQuery';
import type { KwdbVariableQuery } from '../src/types';

describe('parseVariableQuery', () => {
  it('parses tag_values with table and column', () => {
    expect(parseVariableQuery('tag_values(equipment_telemetry, device_id)')).toEqual({
      queryType: 'tagValues',
      table: 'equipment_telemetry',
      column: 'device_id',
    });
  });

  it('tolerates whitespace and quoted identifiers', () => {
    expect(parseVariableQuery(`tag_values( "equipment_telemetry" , 'device_id' )`)).toEqual({
      queryType: 'tagValues',
      table: 'equipment_telemetry',
      column: 'device_id',
    });
  });

  it('parses tables() case-insensitively and ignores a trailing semicolon', () => {
    expect(parseVariableQuery('Tables();')).toEqual({ queryType: 'tables' });
  });

  it('parses columns(table)', () => {
    expect(parseVariableQuery('columns(equipment_telemetry)')).toEqual({
      queryType: 'columns',
      table: 'equipment_telemetry',
    });
  });

  it('accepts an already structured query', () => {
    const query: KwdbVariableQuery = { queryType: 'tagValues', table: 'telemetry', column: 'device_id' };
    expect(parseVariableQuery(query)).toBe(query);
  });

  it('rejects malformed queries', () => {
    expect(() => parseVariableQuery('select 1')).toThrow();
    expect(() => parseVariableQuery('tag_values(only_table)')).toThrow();
    expect(() => parseVariableQuery('columns()')).toThrow();
    expect(() => parseVariableQuery('tables(something)')).toThrow();
    expect(() => parseVariableQuery(undefined)).toThrow();
  });

  it('rejects structured queries with missing fields', () => {
    expect(() => parseVariableQuery({ queryType: 'tagValues', table: 'telemetry' })).toThrow();
    expect(() => parseVariableQuery({ queryType: 'columns' })).toThrow();
  });
});

describe('buildVariableQueryString', () => {
  it('renders the canonical string for each query type', () => {
    expect(buildVariableQueryString({ queryType: 'tables' })).toBe('tables()');
    expect(buildVariableQueryString({ queryType: 'columns', table: 't1' })).toBe('columns(t1)');
    expect(buildVariableQueryString({ queryType: 'tagValues', table: 't1', column: 'c1' })).toBe('tag_values(t1, c1)');
  });

  it('round-trips through parseVariableQuery', () => {
    const query: KwdbVariableQuery = { queryType: 'tagValues', table: 'equipment_telemetry', column: 'device_id' };
    expect(parseVariableQuery(buildVariableQueryString(query))).toEqual(query);
  });
});
