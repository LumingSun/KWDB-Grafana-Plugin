import React from 'react';
import { IconButton, Input, Select, Stack } from '@grafana/ui';

import type { ColumnInfo, MetricSpec } from '../../../types';
import { ColumnPicker } from './ColumnPicker';

const AGGREGATIONS: Array<MetricSpec['aggregation']> = ['avg', 'sum', 'count', 'min', 'max', 'stddev', 'none'];

interface Props {
  index: number;
  columns: ColumnInfo[];
  metric: MetricSpec;
  onChange: (metric: MetricSpec) => void;
  onRemove?: () => void;
  hideAggregation?: boolean;
}

export function MetricRow({ index, columns, metric, onChange, onRemove, hideAggregation }: Props) {
  return (
    <Stack gap={1}>
      <ColumnPicker
        kind="metric"
        columns={columns}
        value={metric.column ? [metric.column] : []}
        onChange={(value) => onChange({ ...metric, column: value[0] ?? '' })}
        ariaLabel={`Metric column ${index}`}
        width={20}
      />
      {!hideAggregation && (
        <Select<MetricSpec['aggregation']>
          aria-label={`Aggregation ${index}`}
          options={AGGREGATIONS.map((aggregation) => ({ label: aggregation, value: aggregation }))}
          value={{ label: metric.aggregation, value: metric.aggregation }}
          onChange={(item) => onChange({ ...metric, aggregation: item.value ?? 'avg' })}
          menuShouldPortal={false}
          width={16}
        />
      )}
      <Input
        aria-label={`Alias ${index}`}
        value={metric.alias ?? ''}
        onChange={(event) => onChange({ ...metric, alias: event.currentTarget.value })}
        placeholder="Alias (optional)"
        width={20}
      />
      {onRemove && <IconButton name="trash-alt" tooltip="Remove metric" onClick={onRemove} />}
    </Stack>
  );
}
