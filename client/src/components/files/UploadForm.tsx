import { useRef, useState } from "react";
import { Button } from "../../UI/Button";
import { uploadFile } from "../../http/api";

type Props = {
  onUploadSuccess: () => void;
};

export function UploadForm({ onUploadSuccess }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState<{ fileId: string } | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const file = inputRef.current?.files?.[0];
    if (!file) return;

    setUploading(true);
    setResult(null);
    setError(null);

    try {
      const data = await uploadFile(file);
      setResult(data);
      if (inputRef.current) inputRef.current.value = "";
      onUploadSuccess();
    } catch {
      setError("Upload failed. Please try again.");
    } finally {
      setUploading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ marginBottom: "1rem" }}>
      <input ref={inputRef} type="file" required />
      {" "}
      <Button type="submit" disabled={uploading}>
        {uploading ? "Uploading..." : "Upload"}
      </Button>
      {result && (
        <p style={{ color: "green" }}>
          File queued for sanitization. ID: <code>{result.fileId}</code> — pending sanitization
        </p>
      )}
      {error && <p style={{ color: "red" }}>{error}</p>}
    </form>
  );
}
