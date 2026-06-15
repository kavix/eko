"use client";
import { useEffect, useRef } from "react";
import type { Snapshot } from "@/lib/data";
import styles from "./TimelineSlider.module.css";

interface Props {
  snapshots: Snapshot[];
  selected: Snapshot | null;
  onSelect: (s: Snapshot) => void;
}

export default function TimelineSlider({ snapshots, selected, onSelect }: Props) {
  const trackRef = useRef<HTMLDivElement>(null);

  if (snapshots.length === 0) return null;

  const selectedIndex = selected ? snapshots.findIndex((s) => s.id === selected.id) : -1;

  const handleClick = (s: Snapshot) => onSelect(s);

  const handleTrackClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!trackRef.current || snapshots.length < 2) return;
    const rect = trackRef.current.getBoundingClientRect();
    const ratio = (e.clientX - rect.left) / rect.width;
    const idx = Math.round(ratio * (snapshots.length - 1));
    const clamped = Math.max(0, Math.min(snapshots.length - 1, idx));
    onSelect(snapshots[clamped]);
  };

  const progressPct =
    snapshots.length > 1
      ? (selectedIndex / (snapshots.length - 1)) * 100
      : 0;

  return (
    <div className={styles.wrap}>
      <span className={styles.label}>◀</span>

      <div className={styles.track} ref={trackRef} onClick={handleTrackClick} id="timeline-slider">
        {/* Filled rail */}
        <div className={styles.rail}>
          <div
            className={styles.fill}
            style={{ width: `${progressPct}%` }}
          />
        </div>

        {/* Dots */}
        {snapshots.map((s, i) => {
          const pct = snapshots.length > 1 ? (i / (snapshots.length - 1)) * 100 : 50;
          const isActive = selected?.id === s.id;
          return (
            <button
              key={s.id}
              id={`slider-dot-${s.id}`}
              className={`${styles.dot} ${isActive ? styles.dotActive : ""}`}
              style={{ left: `${pct}%` }}
              onClick={(e) => { e.stopPropagation(); handleClick(s); }}
              title={s.prompt}
            />
          );
        })}
      </div>

      <span className={styles.label}>▶</span>

      {selected && (
        <span className={styles.badge}>
          {selected.timestamp}
          &nbsp;·&nbsp;#{selected.id}
        </span>
      )}
    </div>
  );
}
