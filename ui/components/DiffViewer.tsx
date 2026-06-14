"use client";
import { useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
import styles from "./DiffViewer.module.css";
import type { Snapshot } from "./SnapshotCard";

// Load Monaco only on the client side
const MonacoDiffEditor = dynamic(
  () => import("@monaco-editor/react").then((m) => m.DiffEditor),
  { ssr: false, loading: () => <div className={styles.loading}><div className={styles.spinner} /></div> }
);

interface DiffFile {
  name: string;
  original: string;
  modified: string;
}

interface Props {
  fromSnapshot: Snapshot;
  toSnapshot: Snapshot;
  onClose: () => void;
}

export default function DiffViewer({ fromSnapshot, toSnapshot, onClose }: Props) {
  const [files, setFiles] = useState<DiffFile[]>([]);
  const [activeFile, setActiveFile] = useState<DiffFile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    setLoading(true);
    setError("");
    fetch(`/api/diff?from=${fromSnapshot.id}&to=${toSnapshot.id}`)
      .then((r) => r.json())
      .then((data: DiffFile[]) => {
        setFiles(data);
        setActiveFile(data[0] ?? null);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  }, [fromSnapshot.id, toSnapshot.id]);

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <h2 className={styles.heading}>Diff</h2>
            <span className={styles.from}>#{fromSnapshot.id}</span>
            <span className={styles.arrow}>→</span>
            <span className={styles.to}>#{toSnapshot.id}</span>
          </div>
          <button className="btn btn-ghost" id="diff-close" onClick={onClose}>✕ Close</button>
        </div>

        <div className={styles.body}>
          {/* File list */}
          {!loading && files.length > 0 && (
            <div className={styles.sidebar}>
              <p className={styles.sidebarLabel}>{files.length} changed files</p>
              <ul className={styles.fileList}>
                {files.map((f) => (
                  <li
                    key={f.name}
                    className={`${styles.fileItem} ${activeFile?.name === f.name ? styles.fileActive : ""}`}
                    onClick={() => setActiveFile(f)}
                    id={`diff-file-${f.name.replace(/\//g, "-")}`}
                  >
                    <span className={styles.fileIcon}>
                      {f.original === "" ? "✚" : f.modified === "" ? "✕" : "~"}
                    </span>
                    <span className={styles.fileName}>{f.name}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Monaco diff */}
          <div className={styles.editor}>
            {loading && <div className={styles.loading}><div className={styles.spinner} /></div>}
            {error && <div className={styles.error}>⚠ {error}</div>}
            {!loading && !error && files.length === 0 && (
              <div className={styles.noChanges}>
                <span>✓</span>
                <p>No differences between snapshots</p>
              </div>
            )}
            {!loading && !error && activeFile && (
              <div className={styles.monacoWrap}>
                <MonacoDiffEditor
                  height="100%"
                  language={getLanguage(activeFile.name)}
                  original={activeFile.original}
                  modified={activeFile.modified}
                  theme="vs-dark"
                  options={{
                    readOnly: true,
                    renderSideBySide: true,
                    minimap: { enabled: false },
                    fontSize: 12,
                    lineNumbers: "on",
                    scrollBeyondLastLine: false,
                    wordWrap: "on",
                  }}
                />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function getLanguage(name: string): string {
  if (name.endsWith(".go")) return "go";
  if (name.endsWith(".ts") || name.endsWith(".tsx")) return "typescript";
  if (name.endsWith(".js") || name.endsWith(".jsx")) return "javascript";
  if (name.endsWith(".json")) return "json";
  if (name.endsWith(".md")) return "markdown";
  if (name.endsWith(".css")) return "css";
  if (name.endsWith(".html")) return "html";
  if (name.endsWith(".py")) return "python";
  return "plaintext";
}
