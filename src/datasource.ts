import { CoreApp, DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { ColumnInfo, DEFAULT_QUERY, KwdbDataSourceOptions, KwdbQuery } from './types';

export class KwdbDataSource extends DataSourceWithBackend<KwdbQuery, KwdbDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<KwdbDataSourceOptions>) {
    super(instanceSettings);
  }

  getTables(): Promise<string[]> {
    return this.getResource('/tables');
  }

  getColumns(table: string): Promise<ColumnInfo[]> {
    return this.getResource(`/columns?table=${encodeURIComponent(table)}`);
  }

  getDefaultQuery(_: CoreApp): Partial<KwdbQuery> {
    return { ...DEFAULT_QUERY };
  }

  filterQuery(query: KwdbQuery): boolean {
    if (query.mode === 'raw') {
      return Boolean(query.rawSql?.trim());
    }
    return Boolean(query.table);
  }
}
