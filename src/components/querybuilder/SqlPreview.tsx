import React, { useRef, useState } from 'react';
import { Button, CodeEditor } from '@grafana/ui';

interface Props {
  sql: string;
}

const MIN_HEIGHT = 140;
const MAX_HEIGHT = 600;

export function SqlPreview({ sql }: Props) {
  const [copied, setCopied] = useState(false);
  const [height, setHeight] = useState(MIN_HEIGHT);
  const startYRef = useRef(0);
  const startHeightRef = useRef(MIN_HEIGHT);

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

  const startResize = (event: React.MouseEvent<HTMLDivElement>) => {
    event.preventDefault();
    event.stopPropagation();
    startYRef.current = event.clientY;
    startHeightRef.current = height;

    const handleMove = (moveEvent: MouseEvent) => {
      const nextHeight = Math.min(
        MAX_HEIGHT,
        Math.max(MIN_HEIGHT, startHeightRef.current + moveEvent.clientY - startYRef.current)
      );
      setHeight(nextHeight);
    };
    const handleUp = () => {
      window.removeEventListener('mousemove', handleMove);
      window.removeEventListener('mouseup', handleUp);
    };
    window.addEventListener('mousemove', handleMove);
    window.addEventListener('mouseup', handleUp);
  };

  return (
    <div data-testid="sql-preview" style={{ width: '100%', minWidth: 0, boxSizing: 'border-box' }}>
      <Button icon="copy" size="sm" variant="secondary" aria-label="Copy SQL" onClick={copySql}>
        {copied ? 'Copied' : 'Copy'}
      </Button>
      <div data-testid="sql-preview-resizable" style={{ width: '100%' }}>
        <CodeEditor value={sql} language="sql" width="100%" height={height} readOnly showLineNumbers={false} />
      </div>
      <div
        data-testid="sql-preview-resize-handle"
        onMouseDown={startResize}
        style={{
          height: 10,
          cursor: 'ns-resize',
          position: 'relative',
          zIndex: 1,
          pointerEvents: 'auto',
          userSelect: 'none',
          touchAction: 'none',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderTop: '1px solid var(--border-color)',
        }}
      >
        <div style={{ width: 32, height: 3, borderRadius: 2, background: 'var(--text-muted)' }} />
      </div>
    </div>
  );
}
