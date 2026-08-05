import React from 'react';
import { CodeEditor } from '@grafana/ui';

interface Props {
  sql: string;
}

export function SqlPreview({ sql }: Props) {
  return (
    <div data-testid="sql-preview">
      <CodeEditor value={sql} language="sql" height={140} readOnly showLineNumbers={false} />
    </div>
  );
}
