import { useEffect, useRef } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getFiles } from "../../http/api";
import { FileList } from "./FileList";
import { UploadForm } from "./UploadForm";

const POLL_INTERVAL = 5000;

export function FileManager() {
  const queryClient = useQueryClient();
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { data: files = [], isLoading, error } = useQuery({
    queryKey: ["files"],
    queryFn: getFiles,
    refetchOnWindowFocus: false,
  });

  // Poll every 5s while any file is pending
  useEffect(() => {
    const hasPending = files.some((f) => f.status === "pending");

    if (hasPending) {
      if (!intervalRef.current) {
        intervalRef.current = setInterval(() => {
          queryClient.invalidateQueries({ queryKey: ["files"] });
        }, POLL_INTERVAL);
      }
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [files, queryClient]);

  const handleUploadSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ["files"] });
  };

  return (
    <div>
      <h2>My Files</h2>
      <UploadForm onUploadSuccess={handleUploadSuccess} />
      {isLoading && <p>Loading files...</p>}
      {error && <p style={{ color: "red" }}>Failed to load files.</p>}
      {!isLoading && <FileList files={files} />}
    </div>
  );
}
