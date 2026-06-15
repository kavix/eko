"use client";
import type { Snapshot } from "@/lib/data";
import styles from "./SnapshotCard.module.css";

interface Props {
  snapshot: Snapshot;
  index: number;
  isSelected: boolean;
  onSelect: () => void;
  onDiff: () => void;
  onRestore: () => void;
}

export default function SnapshotCard({
  snapshot, index, isSelected, onSelect, onDiff, onRestore,
}: Props) {
  return (
    <article
      id={`card-${snapshot.id}`}
      className={`${styles.card} ${isSelected ? styles.selected : ""}`}
      onClick={onSelect}
      aria-selected={isSelected}
    >
      {/* Rail */}
      <div className={styles.rail}>
        <div className={styles.node} />
        <div className={styles.line} />
      </div>

      {/* Content */}
      <div className={styles.content}>
        <div className={styles.meta}>
          <time className={styles.time}>{snapshot.timestamp}</time>
          <span className={`tag ${styles.fileCount}`}>
            {snapshot.filesChanged.length} files
          </span>
        </div>

        <h3 className={styles.prompt}>{snapshot.prompt}</h3>
        <p className={styles.summary}>{snapshot.aiSummary}</p>

        <div className={styles.actions}>
          <button
            className="btn"
            id={`view-${snapshot.id}`}
            onClick={(e) => { e.stopPropagation(); onSelect(); }}
          >
            View
          </button>
          <button
            className="btn"
            id={`diff-${snapshot.id}`}
            onClick={(e) => { e.stopPropagation(); onDiff(); }}
          >
            Diff
          </button>
          <button
            className="btn btn-primary"
            id={`restore-${snapshot.id}`}
            onClick={(e) => { e.stopPropagation(); onRestore(); }}
          >
            Restore
          </button>
        </div>
      </div>
    </article>
  );
}
