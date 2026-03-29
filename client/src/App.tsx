import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Login from "./components/auth/Login";
import Register from "./components/auth/Register";
import { Button } from "./UI/Button";
import { FileManager } from "./components/files/FileManager";
import { fetchSession, logoutUser } from "./http/api";

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
      <FileManager />
    </div>
  );
}

export default App;
