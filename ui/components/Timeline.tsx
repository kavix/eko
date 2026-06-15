"use client";
import type { Snapshot } from "@/lib/data";
import SnapshotCard from "./SnapshotCard";
import styles from "./Timeline.module.css";

interface Props {
  projectName: string;
  snapshots: Snapshot[];
  selectedId: string | null;
  onSelect: (s: Snapshot) => void;
  onDiff: (s: Snapshot) => void;
  onRestore: (s: Snapshot) => void;
  onCreateSnapshot: () => void;
}

export default function Timeline({
  projectName, snapshots, selectedId, onSelect, onDiff, onRestore, onCreateSnapshot,
}: Props) {
  return (
    <section className={styles.panel}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <span className="label">Timeline</span>
          <span className={styles.projectName}>{projectName}</span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <button
            onClick={onCreateSnapshot}
            style={{
              background: "#38bdf8",
              color: "#0f172a",
              border: "none",
              borderRadius: "6px",
              padding: "6px 12px",
              fontSize: "12px",
              fontWeight: "bold",
              cursor: "pointer",
            }}
          >
            + New Snapshot
          </button>
          <span className={styles.count}>{snapshots.length} snapshots</span>
        </div>
      </div>

      {/* Snapshot list */}
      <div className={styles.feed} role="list">
        {snapshots.map((s, i) => (
          <SnapshotCard
            key={s.id}
            snapshot={s}
            index={i}
            isSelected={selectedId === s.id}
            onSelect={() => onSelect(s)}
            onDiff={() => onDiff(s)}
            onRestore={() => onRestore(s)}
          />
        ))}
      </div>
    </section>
  );
}
