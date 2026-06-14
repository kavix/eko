"use client";
import { useEffect, useRef, useState } from "react";
import SnapshotCard, { type Snapshot } from "./SnapshotCard";
import styles from "./Timeline.module.css";

interface Props {
  snapshots: Snapshot[];
  selected: Snapshot | null;
  loading: boolean;
  onSelect: (s: Snapshot) => void;
  onDiff: (s: Snapshot) => void;
  onRestore: (s: Snapshot) => void;
}

export default function Timeline({
  snapshots, selected, loading, onSelect, onDiff, onRestore,
}: Props) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [query, setQuery] = useState("");

  const filtered = snapshots.filter((s) =>
    s.message.toLowerCase().includes(query.toLowerCase()) ||
    s.id.toLowerCase().includes(query.toLowerCase())
  );

  // Scroll selected card into view
  useEffect(() => {
    if (selected) {
      const el = document.getElementById(`snapshot-${selected.id}`);
      el?.scrollIntoView({ behavior: "smooth", block: "nearest" });
    }
  }, [selected]);

  return (
    <div className={styles.panel}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.title}>
          <h1 className={styles.heading}>Timeline</h1>
          <span className={`badge badge-purple`}>{snapshots.length} snapshots</span>
        </div>
        <input
          className={styles.search}
          placeholder="Search snapshots…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          id="timeline-search"
        />
      </div>

      {/* Cards */}
      <div className={styles.feed} ref={scrollRef}>
        {loading && (
          <div className={styles.skeletons}>
            {[0, 1, 2].map((i) => (
              <div key={i} className={styles.skeletonCard}>
                <div className={`skeleton ${styles.skeletonDot}`} />
                <div className={styles.skeletonBody}>
                  <div className={`skeleton ${styles.skeletonLine}`} style={{ width: "40%" }} />
                  <div className={`skeleton ${styles.skeletonLine}`} style={{ width: "80%" }} />
                  <div className={`skeleton ${styles.skeletonLine}`} style={{ width: "60%" }} />
                </div>
              </div>
            ))}
          </div>
        )}

        {!loading && filtered.length === 0 && (
          <div className={styles.empty}>
            <span className={styles.emptyIcon}>🌌</span>
            <p>No snapshots yet.</p>
            <p className={styles.emptyHint}>Run <code>eko save</code> to capture a moment.</p>
          </div>
        )}

        {!loading && filtered.map((s, i) => (
          <SnapshotCard
            key={s.id}
            snapshot={s}
            index={i}
            isSelected={selected?.id === s.id}
            onSelect={onSelect}
            onDiff={onDiff}
            onRestore={onRestore}
          />
        ))}
      </div>
    </div>
  );
}
