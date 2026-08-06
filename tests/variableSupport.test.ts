import { DataSourceInstanceSettings, StandardVariableQuery } from '@grafana/data';
import { TemplateSrv } from '@grafana/runtime';

import { KwdbDataSource } from '../src/datasource';
import type { KwdbDataSourceOptions, KwdbQuery } from '../src/types';
import { KwdbVariableSupport } from '../src/variableSupport';

function makeSettings(): DataSourceInstanceSettings<KwdbDataSourceOptions> {
  return {
    type: 'lumingsun-kwdbtsdb-datasource',
    uid: 'kwdb',
    name: 'KWDB TSDB',
    access: 'proxy',
    jsonData: {},
    meta: {},
  } as unknown as DataSourceInstanceSettings<KwdbDataSourceOptions>;
}

function makeTemplateSrv(impl: (sql: string) => string): TemplateSrv {
  return {
    replace: jest.fn(impl),
  } as unknown as TemplateSrv;
}

describe('KwdbVariableSupport', () => {
  it('maps a standard variable query to a raw KWDB query', () => {
    const support = new KwdbVariableSupport();
    const variableQuery: StandardVariableQuery = {
      refId: 'A',
      query: 'SELECT DISTINCT location FROM demo_ts.sensors',
    };

    expect(support.toDataQuery(variableQuery)).toEqual({
      refId: 'A',
      mode: 'raw',
      format: 'table',
      rawSql: 'SELECT DISTINCT location FROM demo_ts.sensors',
    });
  });

  it('uses variable-query as the fallback refId', () => {
    const support = new KwdbVariableSupport();
    const variableQuery: StandardVariableQuery = {
      refId: '',
      query: 'SHOW TABLES FROM demo_ts',
    };

    expect(support.toDataQuery(variableQuery).refId).toBe('variable-query');
  });
});

describe('KwdbDataSource.applyTemplateVariables', () => {
  it('replaces variables through TemplateSrv with sqlstring formatting', () => {
    const templateSrv = makeTemplateSrv((sql) => sql.replace('$location', "'room-a'"));
    const datasource = new KwdbDataSource(makeSettings(), templateSrv);
    const query: KwdbQuery = {
      refId: 'A',
      mode: 'raw',
      format: 'table',
      rawSql: 'SELECT * FROM demo_ts.sensors WHERE location = $location',
    };

    const result = datasource.applyTemplateVariables(query, {});

    expect(result.rawSql).toBe("SELECT * FROM demo_ts.sensors WHERE location = 'room-a'");
    expect(templateSrv.replace).toHaveBeenCalledWith(
      'SELECT * FROM demo_ts.sensors WHERE location = $location',
      {},
      'sqlstring'
    );
  });

  it('keeps plugin time macros intact for backend expansion', () => {
    const templateSrv = makeTemplateSrv((sql) => sql);
    const datasource = new KwdbDataSource(makeSettings(), templateSrv);
    const query: KwdbQuery = {
      refId: 'A',
      mode: 'raw',
      format: 'time_series',
      rawSql:
        'SELECT $__timeGroup(ts, \'5m\') AS time FROM demo_ts.sensors WHERE $__timeFilter(ts) AND location = $location',
    };

    const result = datasource.applyTemplateVariables(query, {});

    expect(result.rawSql).toBe(
      'SELECT $__timeGroup(ts, \'5m\') AS time FROM demo_ts.sensors WHERE $__timeFilter(ts) AND location = $location'
    );
  });

  it('formats multi-value variables as SQL string literals', () => {
    const templateSrv = makeTemplateSrv((sql) => sql.replace('$location', "'room-a','room-b'"));
    const datasource = new KwdbDataSource(makeSettings(), templateSrv);
    const query: KwdbQuery = {
      refId: 'A',
      mode: 'raw',
      format: 'table',
      rawSql: 'SELECT * FROM demo_ts.sensors WHERE location IN ($location)',
    };

    const result = datasource.applyTemplateVariables(query, {});

    expect(result.rawSql).toBe("SELECT * FROM demo_ts.sensors WHERE location IN ('room-a','room-b')");
  });

  it('returns the query unchanged when rawSql is empty', () => {
    const templateSrv = makeTemplateSrv((sql) => sql);
    const datasource = new KwdbDataSource(makeSettings(), templateSrv);
    const query: KwdbQuery = {
      refId: 'A',
      mode: 'downsampling',
      format: 'time_series',
      rawSql: '',
    };

    expect(datasource.applyTemplateVariables(query, {})).toBe(query);
    expect(templateSrv.replace).not.toHaveBeenCalled();
  });
});
