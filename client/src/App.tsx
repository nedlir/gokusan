import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import Login from "./components/auth/Login";
import Register from "./components/auth/Register";
import { Button } from "./UI/Button";
import { fetchSession, logoutUser, uploadFile, downloadFile } from "./http/api";

function App() {
  const [showRegister, setShowRegister] = useState(false);

  const { data: sessionData, isLoading } = useQuery({
    queryKey: ["session"],
    queryFn: fetchSession,
    retry: false,
    refetchOnWindowFocus: false,
  });

  const isAuthenticated = sessionData?.valid === true;
  const user = sessionData?.user || null;

  const uploadMutation = useMutation({
    mutationFn: uploadFile,
    onSuccess: (data) => {
      alert(`Upload successful! ${data}`);
    },
    onError: (error) => {
      console.error("Upload failed:", error);
      alert("Upload failed");
    },
  });

  const downloadMutation = useMutation({
    mutationFn: downloadFile,
    onSuccess: (data) => {
      alert(`Download successful! ${data}`);
    },
    onError: (error) => {
      console.error("Download failed:", error);
      alert("Download failed");
    },
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated || !user) {
    return (
      <div>
        {showRegister ? (
          <div>
            <Register />
            <p>
              Already have an account?{" "}
              <Button onClick={() => setShowRegister(false)}>Login here</Button>
            </p>
          </div>
        ) : (
          <div>
            <Login />
            <p>
              Don't have an account?{" "}
              <Button onClick={() => setShowRegister(true)}>
                Register here
              </Button>
            </p>
          </div>
        )}
      </div>
    );
  }

  const handleLogout = async () => {
    await logoutUser();
    window.location.reload();
  };

  return (
    <div>
      <div>
        <div>Welcome, {user.name}!</div>
        <p>Role: {user.role}</p>
        <Button onClick={handleLogout}>Logout</Button>
      </div>

      <div>
        <Button
          onClick={() => uploadMutation.mutate()}
          disabled={uploadMutation.isPending}
        >
          {uploadMutation.isPending ? "Uploading..." : "Upload"}
        </Button>
        <Button
          onClick={() => downloadMutation.mutate()}
          disabled={downloadMutation.isPending}
        >
          {downloadMutation.isPending ? "Downloading..." : "Download"}
        </Button>
      </div>
    </div>
  );
}

export default App;
