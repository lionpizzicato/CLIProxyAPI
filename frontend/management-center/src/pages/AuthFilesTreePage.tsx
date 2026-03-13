import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { SecondaryScreenShell } from '@/components/common/SecondaryScreenShell';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { EmptyState } from '@/components/ui/EmptyState';
import { Input } from '@/components/ui/Input';
import { IconChevronDown, IconChevronRight, IconFileText, IconRefreshCw, IconBookOpen } from '@/components/ui/icons';
import { useEdgeSwipeBack } from '@/hooks/useEdgeSwipeBack';
import { useAuthStore } from '@/stores';
import { authFilesApi } from '@/services/api';
import type { AuthFileTreeNode } from '@/types';
import styles from './AuthFilesTreePage.module.scss';

type TreeRow = {
  node: AuthFileTreeNode;
  depth: number;
};

const ROOT_KEY = '__root__';

const toKey = (node: AuthFileTreeNode) => node.path || ROOT_KEY;

const formatSize = (value?: number) => {
  if (!value || value <= 0) return '';
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`;
  return `${(value / (1024 * 1024 * 1024)).toFixed(1)} GB`;
};

const formatModtime = (value?: string | number) => {
  if (value === undefined || value === null || value === '') return '';
  const date = typeof value === 'number' ? new Date(value) : new Date(String(value));
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString();
};

const collectDirPaths = (node: AuthFileTreeNode | null | undefined, paths: Set<string>) => {
  if (!node) return;
  if (node.type === 'dir') {
    paths.add(toKey(node));
    node.children?.forEach((child) => collectDirPaths(child, paths));
  }
};

const filterTree = (node: AuthFileTreeNode | null | undefined, query: string): AuthFileTreeNode | null => {
  if (!node) return null;
  const normalized = query.trim().toLowerCase();
  if (!normalized) return node;

  const nameMatch = node.name.toLowerCase().includes(normalized);
  const pathMatch = (node.path || '').toLowerCase().includes(normalized);
  const isMatch = nameMatch || pathMatch;

  if (node.type !== 'dir') {
    return isMatch ? node : null;
  }

  const filteredChildren = (node.children ?? [])
    .map((child) => filterTree(child, normalized))
    .filter((child): child is AuthFileTreeNode => Boolean(child));

  if (!isMatch && filteredChildren.length === 0) {
    return null;
  }

  return {
    ...node,
    children: filteredChildren,
  };
};

export function AuthFilesTreePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const connectionStatus = useAuthStore((state) => state.connectionStatus);
  const disableControls = connectionStatus !== 'connected';

  const [rootPath, setRootPath] = useState('');
  const [tree, setTree] = useState<AuthFileTreeNode | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('');
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set([ROOT_KEY]));

  const loadTree = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await authFilesApi.getTree();
      setRootPath(result?.root ?? '');
      setTree(result?.tree ?? null);
      setExpanded(new Set([ROOT_KEY]));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : '';
      setError(message || t('notification.load_failed'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadTree();
  }, [loadTree]);

  const handleBack = useCallback(() => {
    navigate('/auth-files', { replace: true });
  }, [navigate]);

  const swipeRef = useEdgeSwipeBack({ onBack: handleBack });

  const filteredTree = useMemo(() => filterTree(tree, filter), [tree, filter]);

  const rows = useMemo(() => {
    const output: TreeRow[] = [];
    if (!filteredTree) return output;
    const autoExpand = filter.trim() !== '';

    const walk = (node: AuthFileTreeNode, depth: number) => {
      output.push({ node, depth });
      if (node.type !== 'dir') return;
      const key = toKey(node);
      const shouldExpand = autoExpand || expanded.has(key);
      if (!shouldExpand) return;
      node.children?.forEach((child) => walk(child, depth + 1));
    };

    walk(filteredTree, 0);
    return output;
  }, [expanded, filteredTree, filter]);

  const totalDirs = useMemo(() => {
    if (!tree) return 0;
    const paths = new Set<string>();
    collectDirPaths(tree, paths);
    return paths.size;
  }, [tree]);

  const handleExpandAll = useCallback(() => {
    if (!tree) return;
    const paths = new Set<string>();
    collectDirPaths(tree, paths);
    if (!paths.has(ROOT_KEY)) paths.add(ROOT_KEY);
    setExpanded(paths);
  }, [tree]);

  const handleCollapseAll = useCallback(() => {
    setExpanded(new Set([ROOT_KEY]));
  }, []);

  const toggleExpanded = (node: AuthFileTreeNode) => {
    if (node.type !== 'dir') return;
    const key = toKey(node);
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  return (
    <SecondaryScreenShell
      ref={swipeRef}
      title={t('auth_files.tree_title', { defaultValue: 'Auth File Tree' })}
      onBack={handleBack}
      backLabel={t('common.back')}
      backAriaLabel={t('common.back')}
      isLoading={loading}
      loadingLabel={t('common.loading')}
      contentClassName={styles.pageContent}
      rightAction={
        <Button variant="secondary" size="sm" onClick={loadTree} disabled={loading}>
          <IconRefreshCw size={16} />
          {t('common.refresh')}
        </Button>
      }
    >
      <Card className={styles.infoCard}>
        <div className={styles.infoHeader}>
          <div className={styles.infoTitle}>{t('auth_files.tree_root', { defaultValue: 'Auth Dir' })}</div>
          <div className={styles.infoMeta}>
            {totalDirs > 0 && (
              <span>
                {t('auth_files.tree_dirs', { defaultValue: '{{count}} folders', count: totalDirs })}
              </span>
            )}
          </div>
        </div>
        <div className={styles.rootPath}>{rootPath || t('common.unknown')}</div>
        <div className={styles.toolbar}>
          <Input
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder={t('auth_files.tree_search_placeholder', { defaultValue: 'Search file or folder...' })}
            disabled={disableControls || loading}
            className={styles.searchInput}
          />
          <Button
            variant="secondary"
            size="sm"
            onClick={handleExpandAll}
            disabled={!tree || filter.trim() !== ''}
          >
            {t('auth_files.tree_expand_all', { defaultValue: 'Expand all' })}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleCollapseAll}
            disabled={!tree || filter.trim() !== ''}
          >
            {t('auth_files.tree_collapse_all', { defaultValue: 'Collapse all' })}
          </Button>
        </div>
      </Card>

      <Card className={styles.treeCard}>
        {error && <div className={styles.errorBox}>{error}</div>}
        {!error && rows.length === 0 ? (
          <EmptyState
            title={t('auth_files.tree_empty_title', { defaultValue: 'No auth files found' })}
            description={t('auth_files.tree_empty_desc', { defaultValue: 'Upload auth files to see them here.' })}
          />
        ) : (
          <div className={styles.treeBody}>
            {rows.map(({ node, depth }) => {
              const isDir = node.type === 'dir';
              const key = toKey(node);
              const hasChildren = (node.children?.length ?? 0) > 0;
              const expandedState = expanded.has(key);
              const pathLabel = node.path || '.';
              return (
                <div key={key || node.name} className={styles.treeRow} style={{ paddingLeft: depth * 18 }}>
                  {isDir ? (
                    <button
                      type="button"
                      className={styles.expandButton}
                      onClick={() => toggleExpanded(node)}
                      disabled={!hasChildren || filter.trim() !== ''}
                      aria-label={expandedState ? 'collapse' : 'expand'}
                    >
                      {expandedState ? <IconChevronDown size={16} /> : <IconChevronRight size={16} />}
                    </button>
                  ) : (
                    <span className={styles.expandPlaceholder} />
                  )}
                  <span className={styles.nodeIcon}>
                    {isDir ? <IconBookOpen size={16} /> : <IconFileText size={16} />}
                  </span>
                  <span className={styles.nodeName}>{node.name}</span>
                  <span className={styles.nodePath}>{pathLabel}</span>
                  {!isDir && (
                    <span className={styles.nodeMeta}>
                      {node.size ? formatSize(node.size) : ''}
                      {node.modtime ? ` | ${formatModtime(node.modtime)}` : ''}
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </Card>
    </SecondaryScreenShell>
  );
}
