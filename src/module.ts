import { DataSourcePlugin } from '@grafana/data';

import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { KwdbDataSource } from './datasource';
import { KwdbDataSourceOptions, KwdbQuery, KwdbSecureJsonData } from './types';

export const plugin = new DataSourcePlugin<KwdbDataSource, KwdbQuery, KwdbDataSourceOptions, KwdbSecureJsonData>(
  KwdbDataSource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
