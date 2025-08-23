import { useState, useEffect } from "react";
import Login from "./auth/Login";
import Register from "./auth/Register";

interface User {
  name: string;
  role: string;
}

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticatedFromServer, setIsAuthenticatedFromServer] =
    useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [showRegister, setShowRegister] = useState(false);

  useEffect(() => {
    setIsLoading(false);
  }, []);

  const callUpload = async () => {
    try {
      const res = await fetch("http://localhost:8000/upload", {
        method: "GET",
        credentials: "include",
      });

      if (res.status === 401) {
        setUser(null);
        setIsAuthenticatedFromServer(false);
        alert("Please login to access this feature.");
        return;
      }

      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }

      const text = await res.text();
      alert(text);
    } catch (error) {
      console.error("Upload failed:", error);
      if (error instanceof TypeError && error.message.includes("fetch")) {
        alert(
          "Cannot connect to server. Please check if the services are running."
        );
      } else {
        alert("Upload failed");
      }
    }
  };

  const callDownload = async () => {
    try {
      const res = await fetch("http://localhost:8000/download", {
        method: "GET",
        credentials: "include",
      });

      if (res.status === 401) {
        setUser(null);
        setIsAuthenticatedFromServer(false);
        alert("Please login to access this feature.");
        return;
      }

      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }

      const text = await res.text();
      alert(text);
    } catch (error) {
      console.error("Download failed:", error);
      if (error instanceof TypeError && error.message.includes("fetch")) {
        alert(
          "Cannot connect to server. Please check if the services are running."
        );
      } else {
        alert("Download failed");
      }
    }
  };

  const handleLogin = (userData: User) => {
    setUser(userData);
    setIsAuthenticatedFromServer(true);
  };

  const handleLogout = async () => {
    try {
      await fetch("http://localhost:8000/auth/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch (error) {
      console.error("Logout request failed:", error);
    } finally {
      setUser(null);
      setIsAuthenticatedFromServer(false);
      setShowRegister(false);
    }
  };

  const handleRegister = () => {
    setShowRegister(false);
  };

  if (isLoading) {
    return (
      <div>
        <div>Loading...</div>
      </div>
    );
  }

  if (!isAuthenticatedFromServer || !user) {
    return (
      <div>
        {showRegister ? (
          <div>
            <Register onRegister={handleRegister} />
            <p>
              Already have an account?{" "}
              <button onClick={() => setShowRegister(false)}>Login here</button>
            </p>
          </div>
        ) : (
          <div>
            <Login onLogin={handleLogin} />
            <p>
              Don't have an account?{" "}
              <button onClick={() => setShowRegister(true)}>
                Register here
              </button>
            </p>
          </div>
        )}
      </div>
    );
  }

  return (
    <div>
      <div>
        <div>Welcome, {user.name}!</div>
        <p>
          Role: <strong>{user.role}</strong>
        </p>
        <button onClick={handleLogout}>Logout</button>
      </div>

      <div>
        <button onClick={callUpload}>Upload</button>
        <button onClick={callDownload}>Download</button>
      </div>
    </div>
  );
}

export default App;
