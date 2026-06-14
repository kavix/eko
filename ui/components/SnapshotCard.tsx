"use client";
import { useState } from "react";
import styles from "./SnapshotCard.module.css";

export interface Snapshot {
  id: string;
  message: string;
  path: string;
  createdAt: string;
  filesChanged: number;
}

interface Props {
  snapshot: Snapshot;
  index: number;
  isSelected: boolean;
  onSelect: (s: Snapshot) => void;
  onDiff: (s: Snapshot) => void;
  onRestore: (s: Snapshot) => void;
}

export default function SnapshotCard({
  snapshot, index, isSelected, onSelect, onDiff, onRestore,
}: Props) {
  const [hovered, setHovered] = useState(false);

  const time = new Date(snapshot.createdAt).toLocaleTimeString([], {
    hour: "2-digit", minute: "2-digit",
  });

  return (
    <div
      id={`snapshot-${snapshot.id}`}
      className={`${styles.card} ${isSelected ? styles.selected : ""}`}
      style={{ animationDelay: `${index * 40}ms` }}
      onClick={() => onSelect(snapshot)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {/* Timeline connector */}
      <div className={styles.connector}>
        <div className={`${styles.dot} ${isSelected ? styles.dotActive : ""}`} />
        <div className={styles.line} />
      </div>

      {/* Card body */}
      <div className={styles.body}>
        <div className={styles.header}>
          <span className={styles.time}>{time}</span>
          <span className={`badge badge-purple ${styles.files}`}>
            📄 {snapshot.filesChanged} files
          </span>
        </div>

        <p className={styles.message}>{snapshot.message}</p>

        <div className={styles.idRow}>
          <code className={styles.idChip}>#{snapshot.id}</code>
        </div>

        {/* Actions — slide in on hover/select */}
        <div className={`${styles.actions} ${(hovered || isSelected) ? styles.actionsVisible : ""}`}>
          <button
            className="btn btn-ghost"
            id={`view-${snapshot.id}`}
            onClick={(e) => { e.stopPropagation(); onSelect(snapshot); }}
          >
            View
          </button>
          <button
            className="btn btn-ghost"
            id={`diff-${snapshot.id}`}
            onClick={(e) => { e.stopPropagation(); onDiff(snapshot); }}
          >
            Diff
          </button>
          <button
            className="btn btn-primary"
            id={`restore-${snapshot.id}`}
            onClick={(e) => { e.stopPropagation(); onRestore(snapshot); }}
          >
            Restore
          </button>
        </div>
      </div>
    </div>
  );
}
