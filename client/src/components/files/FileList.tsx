import React, { useState } from "react";
import type { File as UserFile } from "../../types/File";
import { Button } from "../../UI/Button";
import { downloadFile, shareFile } from "../../http/api";

type ShareState = {
  fileId: string;
  url: string | null;
  ttl: number;
  loading: boolean;
  error: string | null;
};

type Props = {
  files: UserFile[];
};

const TTL_OPTIONS = [
  { label: "1 hour", seconds: 3600 },
  { label: "6 hours", seconds: 21600 },
  { label: "24 hours", seconds: 86400 },
  { label: "7 days", seconds: 604800 },
];

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function FileList({ files }: Props) {
  const [shareState, setShareState] = useState<ShareState | null>(null);
  const [downloadError, setDownloadError] = useState<string | null>(null);

  const handleDownload = async (file: UserFile) => {
    setDownloadError(null);
    try {
      const blob = await downloadFile(file.id);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = file.name;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch {
      setDownloadError(`Failed to download ${file.name}`);
    }
  };

  const openSharePicker = (fileId: string) => {
    setShareState({ fileId, url: null, ttl: TTL_OPTIONS[0].seconds, loading: false, error: null });
  };

  const handleShare = async () => {
    if (!shareState) return;
    setShareState((s) => s && { ...s, loading: true, error: null });
    try {
      const { url } = await shareFile(shareState.fileId, shareState.ttl);
      setShareState((s) => s && { ...s, url, loading: false });
    } catch {
      setShareState((s) => s && { ...s, loading: false, error: "Failed to create share link" });
    }
  };

  if (files.length === 0) {
    return <p>No files yet. Upload one to get started.</p>;
  }

  return (
    <div>
      {downloadError && <p style={{ color: "red" }}>{downloadError}</p>}
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr>
            <th style={thStyle}>Name</th>
            <th style={thStyle}>Size</th>
            <th style={thStyle}>Type</th>
            <th style={thStyle}>Status</th>
            <th style={thStyle}>Created</th>
            <th style={thStyle}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {files.map((file) => (
            <tr key={file.id}>
              <td style={tdStyle}>{file.name}</td>
              <td style={tdStyle}>{formatSize(file.size)}</td>
              <td style={tdStyle}>{file.mimeType}</td>
              <td style={tdStyle}>{file.status}</td>
              <td style={tdStyle}>{new Date(file.createdAt).toLocaleString()}</td>
              <td style={tdStyle}>
                <Button
                  onClick={() => handleDownload(file)}
                  disabled={file.status !== "ready"}
                >
                  {file.status === "ready" ? "Download" : file.status}
                </Button>
                {" "}
                {file.status === "ready" && (
                  <Button onClick={() => openSharePicker(file.id)}>Share</Button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {shareState && (
        <div style={{ marginTop: "1rem", padding: "1rem", border: "1px solid #ccc" }}>
          <strong>Share file</strong>
          <div style={{ marginTop: "0.5rem" }}>
            <label>
              Expires in:{" "}
              <select
                value={shareState.ttl}
                onChange={(e) =>
                  setShareState((s) => s && { ...s, ttl: Number(e.target.value) })
                }
              >
                {TTL_OPTIONS.map((opt) => (
                  <option key={opt.seconds} value={opt.seconds}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <div style={{ marginTop: "0.5rem" }}>
            <Button onClick={handleShare} disabled={shareState.loading}>
              {shareState.loading ? "Generating..." : "Generate link"}
            </Button>
            {" "}
            <Button onClick={() => setShareState(null)}>Cancel</Button>
          </div>
          {shareState.error && <p style={{ color: "red" }}>{shareState.error}</p>}
          {shareState.url && (
            <p>
              Share URL:{" "}
              <a href={shareState.url} target="_blank" rel="noreferrer">
                {shareState.url}
              </a>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "0.5rem",
  borderBottom: "2px solid #ccc",
};

const tdStyle: React.CSSProperties = {
  padding: "0.5rem",
  borderBottom: "1px solid #eee",
};
