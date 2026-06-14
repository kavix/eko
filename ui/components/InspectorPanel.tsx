"use client";
import styles from "./InspectorPanel.module.css";
import type { Snapshot } from "./SnapshotCard";

interface Props {
  snapshot: Snapshot | null;
  onRestore: (s: Snapshot) => void;
  onDiff: (s: Snapshot) => void;
}

export default function InspectorPanel({ snapshot, onRestore, onDiff }: Props) {
  if (!snapshot) {
    return (
      <div className={styles.panel}>
        <div className={styles.empty}>
          <span className={styles.emptyIcon}>🔍</span>
          <p>Select a snapshot</p>
          <p className={styles.emptyHint}>Click any card to inspect its details</p>
        </div>
      </div>
    );
  }

  const date = new Date(snapshot.createdAt);
  const formatted = date.toLocaleString([], {
    month: "short", day: "numeric",
    hour: "2-digit", minute: "2-digit",
  });

  return (
    <div className={styles.panel} key={snapshot.id}>
      <div className={styles.header}>
        <div className={styles.headerTop}>
          <h2 className={styles.heading}>Inspector</h2>
          <code className={styles.idPill}>#{snapshot.id}</code>
        </div>
        <p className={styles.subheading}>{formatted}</p>
      </div>

      <div className={styles.body}>
        {/* Prompt */}
        <section className={styles.section}>
          <label className={styles.sectionLabel}>Prompt</label>
          <div className={styles.promptBox}>
            <span className={styles.promptIcon}>💬</span>
            <p className={styles.promptText}>{snapshot.message}</p>
          </div>
        </section>

        {/* Model (static for MVP) */}
        <section className={styles.section}>
          <label className={styles.sectionLabel}>Model</label>
          <div className={styles.modelRow}>
            <span className={styles.modelDot} />
            <span className={styles.modelName}>Claude Sonnet 4.6</span>
          </div>
        </section>

        {/* AI Summary */}
        <section className={styles.section}>
          <label className={styles.sectionLabel}>AI Summary</label>
          <p className={styles.summary}>
            Snapshot captured {snapshot.filesChanged} file{snapshot.filesChanged !== 1 ? "s" : ""}.
            Stored at <code className={styles.code}>{snapshot.path}</code>.
          </p>
        </section>

        {/* Files */}
        <section className={styles.section}>
          <label className={styles.sectionLabel}>Files Changed</label>
          <div className={`badge badge-green ${styles.filesBadge}`}>
            📁 {snapshot.filesChanged} files
          </div>
        </section>
      </div>

      {/* Actions */}
      <div className={styles.footer}>
        <button
          className="btn btn-ghost"
          id={`inspector-diff-${snapshot.id}`}
          onClick={() => onDiff(snapshot)}
          style={{ flex: 1 }}
        >
          Compare Diff
        </button>
        <button
          className="btn btn-primary"
          id={`inspector-restore-${snapshot.id}`}
          onClick={() => onRestore(snapshot)}
          style={{ flex: 1 }}
        >
          ↩ Restore
        </button>
      </div>
    </div>
  );
}
