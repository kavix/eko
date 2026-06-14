"use client";
import styles from "./ProjectSidebar.module.css";

const PROJECTS = [
  { id: "eko",          name: "eko",           icon: "⚡", snapshots: 12 },
  { id: "ecommerce-ai", name: "ecommerce-ai",  icon: "🛒", snapshots:  7 },
  { id: "chat-app",     name: "chat-app",      icon: "💬", snapshots:  4 },
  { id: "auth-service", name: "auth-service",  icon: "🔐", snapshots:  9 },
];

interface Props {
  activeProject: string;
  onSelect: (id: string) => void;
}

export default function ProjectSidebar({ activeProject, onSelect }: Props) {
  return (
    <aside className={styles.sidebar}>
      <div className={styles.logo}>
        <span className={styles.logoMark}>eko</span>
        <span className={styles.logoSub}>memory</span>
      </div>

      <div className={styles.section}>
        <span className={styles.sectionLabel}>Projects</span>
        <ul className={styles.list}>
          {PROJECTS.map((p) => (
            <li
              key={p.id}
              id={`project-${p.id}`}
              className={`${styles.item} ${activeProject === p.id ? styles.active : ""}`}
              onClick={() => onSelect(p.id)}
            >
              <span className={styles.icon}>{p.icon}</span>
              <span className={styles.name}>{p.name}</span>
              <span className={styles.count}>{p.snapshots}</span>
            </li>
          ))}
        </ul>
      </div>

      <div className={styles.footer}>
        <div className={styles.footerDot} />
        <span className={styles.footerText}>API connected</span>
      </div>
    </aside>
  );
}
