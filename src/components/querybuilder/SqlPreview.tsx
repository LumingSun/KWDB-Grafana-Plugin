import React, { useState } from 'react';
import { Button, CodeEditor } from '@grafana/ui';

interface Props {
  sql: string;
}

export function SqlPreview({ sql }: Props) {
  const [copied, setCopied] = useState(false);

  const copySql = async () => {
    if (!sql) {
      return;
    }
    try {
      await navigator.clipboard.writeText(sql);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard access can be unavailable in embedded or non-secure contexts.
    }
  };

  return (
    <div data-testid="sql-preview">
      <Button icon="copy" size="sm" variant="secondary" aria-label="Copy SQL" onClick={copySql}>
        {copied ? 'Copied' : 'Copy'}
      </Button>
      <CodeEditor value={sql} language="sql" height={140} readOnly showLineNumbers={false} />
    </div>
  );
}
