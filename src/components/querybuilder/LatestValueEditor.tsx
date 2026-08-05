import React from 'react';
import { Button, InlineField, Select, VerticalGroup } from '@grafana/ui';

import type { KwdbDataSource } from '../../datasource';
import { buildSql } from '../../sqlBuilder';
import type { KwdbQuery, MetricSpec } from '../../types';
import { SqlPreview } from './SqlPreview';
import { ColumnPicker } from './shared/ColumnPicker';
import { MetricRow } from './shared/MetricRow';
import { TableSelector } from './shared/TableSelector';
import { TimeColumnPicker } from './shared/TimeColumnPicker';
import { useTableColumns } from './shared/useTableColumns';

const LATEST_FUNCS: Array<NonNullable<KwdbQuery['latestFunc']>> = ['last', 'last_row', 'first', 'first_row'];

interface Props {
  datasource: KwdbDataSource;
  query: KwdbQuery;
  onChange: (query: KwdbQuery) => void;
}

export function LatestValueEditor({ datasource, query, onChange }: Props) {
  const columns = useTableColumns(datasource, query.table);

  const updateQuery = (patch: Partial<KwdbQuery>) => {
    const next = { ...query, ...patch };
    next.rawSql = buildSql(next);
    onChange(next);
  };

  const metrics: MetricSpec[] = query.metrics?.length ? query.metrics : [{ column: '', aggregation: 'none' }];

  const updateMetric = (index: number, metric: MetricSpec) => {
    updateQuery({ metrics: metrics.map((item, itemIndex) => (itemIndex === index ? metric : item)) });
  };

  const removeMetric = (index: number) => {
    if (metrics.length <= 1) {
      return;
    }
    updateQuery({ metrics: metrics.filter((_, itemIndex) => itemIndex !== index) });
  };

  return (
    <VerticalGroup spacing="md">
      <InlineField label="Table" labelWidth={12}>
        <TableSelector datasource={datasource} value={query.table} onChange={(table) => updateQuery({ table })} />
      </InlineField>
      <InlineField label="Time column" labelWidth={12}>
        <TimeColumnPicker
          columns={columns}
          value={query.timeColumn}
          onChange={(timeColumn) => updateQuery({ timeColumn })}
        />
      </InlineField>
      <InlineField label="Latest function" labelWidth={12}>
        <Select<NonNullable<KwdbQuery['latestFunc']>>
          aria-label="Latest function"
          options={LATEST_FUNCS.map((func) => ({ label: func, value: func }))}
          value={query.latestFunc ?? 'last'}
          onChange={(item) => updateQuery({ latestFunc: item.value ?? 'last' })}
          width={20}
          menuShouldPortal={false}
        />
      </InlineField>
      <InlineField label="Tags" labelWidth={12}>
        <ColumnPicker columns={columns} value={query.tags ?? []} onChange={(tags) => updateQuery({ tags })} />
      </InlineField>
      <InlineField label="Metrics" labelWidth={12}>
        <VerticalGroup spacing="sm">
          {metrics.map((metric, index) => (
            <MetricRow
              key={index}
              index={index}
              columns={columns}
              metric={metric}
              hideAggregation
              onChange={(nextMetric) => updateMetric(index, nextMetric)}
              onRemove={metrics.length > 1 ? () => removeMetric(index) : undefined}
            />
          ))}
          <Button
            icon="plus-circle"
            variant="secondary"
            size="sm"
            onClick={() => updateQuery({ metrics: [...metrics, { column: '', aggregation: 'none' }] })}
          >
            Add metric
          </Button>
        </VerticalGroup>
      </InlineField>
      <SqlPreview sql={query.rawSql ?? ''} />
    </VerticalGroup>
  );
}
