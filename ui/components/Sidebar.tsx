"use client";
import type { Project } from "@/lib/data";
import styles from "./Sidebar.module.css";

interface Props {
  projects: Project[];
  activeId: string;
  onSelect: (id: string) => void;
}

export default function Sidebar({ projects, activeId, onSelect }: Props) {
  return (
    <aside className={styles.sidebar}>
      <div className={styles.header}>
        <span className="label">Projects</span>
      </div>

      <nav className={styles.nav}>
        {projects.map((p) => (
          <button
            key={p.id}
            id={`project-${p.id}`}
            className={`${styles.item} ${activeId === p.id ? styles.active : ""}`}
            onClick={() => onSelect(p.id)}
          >
            <span className={styles.dot} />
            <span className={styles.name}>{p.name}</span>
            <span className={styles.count}>{p.snapshots.length}</span>
          </button>
        ))}
      </nav>

      <div className={styles.footer}>
        <span className={styles.footerText}>vibe timeline</span>
        <span className={styles.footerVersion}>v0.1</span>
      </div>
    </aside>
  );
}
