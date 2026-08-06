import { CoreApp, DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv, TemplateSrv } from '@grafana/runtime';

import { ColumnInfo, DEFAULT_QUERY, KwdbDataSourceOptions, KwdbQuery, TableInfo } from './types';
import { KwdbVariableSupport } from './variableSupport';

export class KwdbDataSource extends DataSourceWithBackend<KwdbQuery, KwdbDataSourceOptions> {
  constructor(
    instanceSettings: DataSourceInstanceSettings<KwdbDataSourceOptions>,
    private readonly templateSrv: TemplateSrv = getTemplateSrv()
  ) {
    super(instanceSettings);
    this.variables = new KwdbVariableSupport();
  }

  getTables(): Promise<TableInfo[]> {
    return this.getResource('/tables');
  }

  getColumns(table: string): Promise<ColumnInfo[]> {
    return this.getResource(`/columns?table=${encodeURIComponent(table)}`);
  }

  getDefaultQuery(_: CoreApp): Partial<KwdbQuery> {
    return { ...DEFAULT_QUERY };
  }

  applyTemplateVariables(query: KwdbQuery, scopedVars: ScopedVars): KwdbQuery {
    if (!query.rawSql) {
      return query;
    }
    return {
      ...query,
      rawSql: this.templateSrv.replace(query.rawSql, scopedVars, 'sqlstring'),
    };
  }

  annotations = {};

  filterQuery(query: KwdbQuery): boolean {
    if (query.mode === 'raw') {
      return Boolean(query.rawSql?.trim());
    }
    return Boolean(query.table);
  }
}
