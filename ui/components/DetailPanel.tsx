"use client";
import type { Snapshot } from "@/lib/data";
import styles from "./DetailPanel.module.css";

interface Props {
  snapshot: Snapshot | null;
  onRestore: (s: Snapshot) => void;
  onDiff: (s: Snapshot) => void;
}

export default function DetailPanel({ snapshot, onRestore, onDiff }: Props) {
  if (!snapshot) {
    return (
      <aside className={styles.panel}>
        <div className={styles.header}>
          <span className="label">Details</span>
        </div>
        <div className={styles.empty}>
          <p className={styles.emptyText}>Select a snapshot to inspect</p>
        </div>
      </aside>
    );
  }

  return (
    <aside className={styles.panel} key={snapshot.id}>
      <div className={styles.header}>
        <span className="label">Details</span>
        <code className={styles.idChip}>#{snapshot.id}</code>
      </div>

      <div className={styles.body}>
        {/* Time and metadata fields */}
        <div className={styles.row}>
          <span className={styles.rowLabel}>Time</span>
          <time className={styles.rowValue}>{snapshot.timestamp}</time>
        </div>

        <div className={styles.row}>
          <span className={styles.rowLabel}>Model</span>
          <span className={styles.rowValue}>{snapshot.model}</span>
        </div>

        <div className={styles.divider} />

        <section className={styles.section}>
          <h4 className={styles.sectionTitle}>Prompt</h4>
          <p className={styles.sectionBody}>{snapshot.prompt}</p>
        </section>

        <section className={styles.section}>
          <h4 className={styles.sectionTitle}>AI Summary</h4>
          <p className={styles.sectionBody}>{snapshot.aiSummary}</p>
        </section>

        <section className={styles.section}>
          <h4 className={styles.sectionTitle}>
            Files Changed
            <span className={styles.fileCountBadge}>{snapshot.filesChanged.length}</span>
          </h4>
          <ul className={styles.fileList}>
            {snapshot.filesChanged.map((f) => (
              <li key={f} className={styles.fileItem}>
                <span className={styles.fileIcon}>—</span>
                <code className={styles.fileName}>{f}</code>
              </li>
            ))}
          </ul>
        </section>
      </div>

      <div className={styles.footer}>
        <button
          className="btn"
          id={`detail-diff-${snapshot.id}`}
          style={{ flex: 1 }}
          onClick={() => onDiff(snapshot)}
        >
          View Diff
        </button>
        <button
          className="btn btn-primary"
          id={`detail-restore-${snapshot.id}`}
          style={{ flex: 1 }}
          onClick={() => onRestore(snapshot)}
        >
          Restore
        </button>
      </div>
    </aside>
  );
}
