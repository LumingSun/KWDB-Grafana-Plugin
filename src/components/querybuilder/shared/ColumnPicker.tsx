import React from 'react';
import { MultiSelect, Select } from '@grafana/ui';

import type { ColumnInfo } from '../../../types';

interface Props {
  columns: ColumnInfo[];
  value: string[];
  onChange: (value: string[]) => void;
  kind?: 'tags' | 'metric';
  ariaLabel?: string;
  placeholder?: string;
  width?: number;
}

export function ColumnPicker({ columns, value, onChange, kind = 'tags', ariaLabel, placeholder, width }: Props) {
  const tagOptions = columns
    .filter((column) => column.isTag)
    .map((column) => ({ label: column.name, value: column.name }));
  const dataOptions = columns
    .filter((column) => !column.isTag)
    .map((column) => ({ label: column.name, value: column.name }));

  if (kind === 'metric') {
    return (
      <Select<string>
        aria-label={ariaLabel ?? 'Metric column'}
        options={dataOptions}
        value={value[0] ? { label: value[0], value: value[0] } : undefined}
        onChange={(item) => onChange(item.value ? [item.value] : [])}
        placeholder={placeholder ?? 'Select metric column'}
        isClearable
        menuShouldPortal={false}
        width={width ?? 20}
      />
    );
  }

  return (
    <MultiSelect<string>
      aria-label={ariaLabel ?? 'Tags'}
      options={[{ label: 'Tag columns', options: tagOptions }]}
      value={value.map((item) => ({ label: item, value: item }))}
      onChange={(items) =>
        onChange(items.map((item) => item.value).filter((item): item is string => Boolean(item)))
      }
      placeholder={placeholder ?? 'Select tags for GROUP BY'}
      isClearable
      menuShouldPortal={false}
      width={width ?? 40}
    />
  );
}
