import type { ColumnInfo, MetricSpec } from '../../../types';

export type MetricAggregation = MetricSpec['aggregation'];

export type ColumnKind = 'time' | 'numeric' | 'string' | 'other';

export const AGGREGATIONS: MetricAggregation[] = ['avg', 'sum', 'count', 'min', 'max', 'stddev', 'none'];

const STRING_AGGREGATIONS: MetricAggregation[] = ['count', 'none'];

const TIME_AGGREGATIONS: MetricAggregation[] = ['count', 'min', 'max', 'none'];

// KWDB DDL types such as FLOAT8, INT4, DECIMAL(18,2), DOUBLE, SMALLINT...
const NUMERIC_TYPE_PATTERN = /float|double|real|decimal|numeric|number|int/i;

// KWDB DDL types such as VARCHAR(32), CHAR(10), TEXT, STRING...
const STRING_TYPE_PATTERN = /char|text|string/i;

export function isTimeColumnType(type: string): boolean {
  const lower = type.toLowerCase();
  return (
    lower.startsWith('timestamp') ||
    lower.startsWith('k_timestamp') ||
    lower.startsWith('time(') ||
    lower === 'time' ||
    lower === 'date'
  );
}

export function getColumnKind(column?: ColumnInfo): ColumnKind {
  if (!column) {
    return 'other';
  }
  // Tags are always grouped as strings even if their declared type looks numeric.
  if (column.isTag) {
    return 'string';
  }
  if (column.isTimeColumn || isTimeColumnType(column.type)) {
    return 'time';
  }
  if (NUMERIC_TYPE_PATTERN.test(column.type)) {
    return 'numeric';
  }
  if (STRING_TYPE_PATTERN.test(column.type)) {
    return 'string';
  }
  return 'other';
}

/**
 * Returns the aggregations that make sense for a column. Unknown columns or
 * unrecognized types keep every option so existing queries never lose choices.
 */
export function getAllowedAggregations(column?: ColumnInfo): MetricAggregation[] {
  switch (getColumnKind(column)) {
    case 'numeric':
      return AGGREGATIONS;
    case 'time':
      return TIME_AGGREGATIONS;
    case 'string':
      return STRING_AGGREGATIONS;
    default:
      return AGGREGATIONS;
  }
}

/** Aggregation list shown for a metric column name within a table's columns. */
export function getAggregationsForColumn(columns: ColumnInfo[], columnName?: string): MetricAggregation[] {
  if (!columnName) {
    return AGGREGATIONS;
  }
  return getAllowedAggregations(columns.find((column) => column.name === columnName));
}
