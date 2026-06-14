"use client";
import { useCallback, useEffect, useState } from "react";
import ProjectSidebar from "@/components/ProjectSidebar";
import Timeline from "@/components/Timeline";
import TimelineSlider from "@/components/TimelineSlider";
import InspectorPanel from "@/components/InspectorPanel";
import DiffViewer from "@/components/DiffViewer";
import type { Snapshot } from "@/components/SnapshotCard";
import styles from "./page.module.css";

export default function Home() {
  const [project, setProject] = useState("eko");
  const [snapshots, setSnapshots] = useState<Snapshot[]>([]);
  const [selected, setSelected] = useState<Snapshot | null>(null);
  const [loading, setLoading] = useState(true);

  // Diff state
  const [diffFrom, setDiffFrom] = useState<Snapshot | null>(null);
  const [diffTo, setDiffTo] = useState<Snapshot | null>(null);
  const [showDiff, setShowDiff] = useState(false);

  // Restore animation state
  const [restoring, setRestoring] = useState(false);

  // ── Fetch snapshots ──────────────────────────────────
  const fetchSnapshots = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/snapshots");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: Snapshot[] = await res.json();
      setSnapshots(data);
      if (data.length > 0 && !selected) setSelected(data[0]);
    } catch {
      // API not running — use rich demo data
      setSnapshots(DEMO_SNAPSHOTS);
      setSelected(DEMO_SNAPSHOTS[0]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchSnapshots(); }, [fetchSnapshots]);

  // ── Restore ──────────────────────────────────────────
  const handleRestore = async (s: Snapshot) => {
    setRestoring(true);
    try {
      await fetch(`/api/snapshots/${s.id}/restore`, { method: "POST" });
    } catch {
      // demo mode — no-op
    }
    // Simulate rewind duration
    setTimeout(() => {
      setRestoring(false);
      setSelected(s);
    }, 1400);
  };

  // ── Diff ─────────────────────────────────────────────
  const handleDiff = (s: Snapshot) => {
    const idx = snapshots.findIndex((x) => x.id === s.id);
    const prev = idx < snapshots.length - 1 ? snapshots[idx + 1] : s;
    setDiffFrom(prev);
    setDiffTo(s);
    setShowDiff(true);
  };

  return (
    <>
      {/* Restore overlay */}
      {restoring && (
        <div className="restore-overlay">
          <div className={styles.rewindIcon}>⏪</div>
          <span className="restore-label">Rewinding timeline…</span>
          <div className="restore-bar-wrap">
            <div className="restore-bar" />
          </div>
        </div>
      )}

      {/* Diff modal */}
      {showDiff && diffFrom && diffTo && (
        <DiffViewer
          fromSnapshot={diffFrom}
          toSnapshot={diffTo}
          onClose={() => setShowDiff(false)}
        />
      )}

      {/* 3-panel shell */}
      <div className="app-shell">
        {/* LEFT */}
        <ProjectSidebar activeProject={project} onSelect={setProject} />

        {/* CENTER */}
        <div className={styles.center}>
          <Timeline
            snapshots={snapshots}
            selected={selected}
            loading={loading}
            onSelect={setSelected}
            onDiff={handleDiff}
            onRestore={handleRestore}
          />
          <TimelineSlider
            snapshots={snapshots}
            selected={selected}
            onSelect={setSelected}
          />
        </div>

        {/* RIGHT */}
        <InspectorPanel
          snapshot={selected}
          onRestore={handleRestore}
          onDiff={handleDiff}
        />
      </div>
    </>
  );
}

// ── Demo data (shown when Go API is not running) ─────
const DEMO_SNAPSHOTS: Snapshot[] = [
  {
    id: "a1b2c3d4",
    message: "Add JWT authentication middleware",
    path: ".vibe/snapshots/a1b2c3d4",
    createdAt: new Date(Date.now() - 5 * 60000).toISOString(),
    filesChanged: 8,
  },
  {
    id: "e5f6a7b8",
    message: "Create user schema and migration",
    path: ".vibe/snapshots/e5f6a7b8",
    createdAt: new Date(Date.now() - 18 * 60000).toISOString(),
    filesChanged: 3,
  },
  {
    id: "c9d0e1f2",
    message: "Scaffold REST API routes for /auth",
    path: ".vibe/snapshots/c9d0e1f2",
    createdAt: new Date(Date.now() - 35 * 60000).toISOString(),
    filesChanged: 5,
  },
  {
    id: "g3h4i5j6",
    message: "Initial project setup — Next.js + Go",
    path: ".vibe/snapshots/g3h4i5j6",
    createdAt: new Date(Date.now() - 72 * 60000).toISOString(),
    filesChanged: 12,
  },
];
