import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type KwdbQueryMode = 'downsampling' | 'gapfill' | 'latest' | 'window' | 'raw';

export type KwdbQueryFormat = 'time_series' | 'table';

export interface MetricSpec {
  column: string;
  aggregation: 'avg' | 'sum' | 'count' | 'min' | 'max' | 'stddev' | 'none';
  alias?: string;
}

export interface KwdbQuery extends DataQuery {
  mode: KwdbQueryMode;
  format: KwdbQueryFormat;
  // Common fields
  table?: string;
  timeColumn?: string;
  tags?: string[];
  metrics?: MetricSpec[];
  // Downsampling / gapfill
  interval?: string;
  // Gapfill
  interpolateMode?: 'linear' | 'PREV' | 'NEXT' | 'constant' | 'NULL';
  // Latest values
  latestFunc?: 'last' | 'last_row' | 'first' | 'first_row';
  // Latest values: split results into one frame per tag combination (default true)
  splitByTag?: boolean;
  // Window / event
  windowType?: 'TIME_WINDOW' | 'SESSION_WINDOW' | 'EVENT_WINDOW' | 'COUNT_WINDOW' | 'STATE_WINDOW';
  windowInterval?: string;
  windowSlide?: string;
  eventStartCond?: string;
  eventEndCond?: string;
  // Raw SQL
  rawSql?: string;
}

export const DEFAULT_QUERY: Partial<KwdbQuery> = {
  mode: 'downsampling',
  format: 'time_series',
  interval: '5m',
  metrics: [{ column: '', aggregation: 'avg' }],
};

export interface KwdbDataSourceOptions extends DataSourceJsonData {
  host?: string;
  port?: number;
  database?: string;
  user?: string;
  sslMode?: 'disable' | 'require' | 'verify-ca' | 'verify-full';
  sslRootCert?: string;
}

export interface KwdbSecureJsonData {
  password?: string;
}

export interface ColumnInfo {
  name: string;
  type: string;
  isTag: boolean;
  isTimeColumn: boolean;
  isPrimaryTag: boolean;
}

export interface TableInfo {
  name: string;
  type: string;
}
