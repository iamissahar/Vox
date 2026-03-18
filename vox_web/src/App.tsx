import React from "react";
import "./App.css";
import { useRouter } from "./hooks/useRouter";
import { useAuth } from "./hooks/useAuth";
import HomePage from "./pages/HomePage";
import LoginPage from "./pages/LoginPage";
import AdminPage from "./pages/AdminPage";
import RoomJoinPage from "./pages/RoomJoinPage";
import RoomPage from "./pages/RoomPage";
import BroadcastPage from "./pages/BroadcastPage";

const App: React.FC = () => {
  const { route, navigate } = useRouter();
  const { user, isLoading, isAuthenticated, setUser, logout } = useAuth();

  // Spinner while checking auth
  if (isLoading) {
    return (
      <div
        style={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <div
          style={{
            width: 28,
            height: 28,
            border: "2px solid #222",
            borderTopColor: "#e8ff5e",
            borderRadius: "50%",
            animation: "spin 0.8s linear infinite",
          }}
        />
      </div>
    );
  }

  // Parse route
  const roomMatch = route.match(/^#\/room\/(.+)$/);
  const roomId = roomMatch ? roomMatch[1] : null;

  const renderPage = () => {
    // Dynamic room route
    if (roomId) {
      return <RoomPage roomId={roomId} navigate={navigate} />;
    }

    switch (route) {
      case "#/":
      case "":
        return (
          <HomePage navigate={navigate} isAuthenticated={isAuthenticated} />
        );

      case "#/login":
        return (
          <LoginPage
            navigate={navigate}
            onLogin={(u) => {
              setUser(u);
              navigate("#/admin");
            }}
          />
        );

      case "#/admin":
        if (!isAuthenticated) {
          navigate("#/login");
          return null;
        }
        return (
          <AdminPage
            navigate={navigate}
            currentUser={user}
            onLogout={() => {
              logout();
              navigate("#/");
            }}
          />
        );

      case "#/room":
        return <RoomJoinPage navigate={navigate} />;

      case "#/broadcast":
        return (
          <BroadcastPage
            hubId={route.replace("#/broadcast/", "")}
            navigate={navigate}
          />
        );

      default:
        return (
          <HomePage navigate={navigate} isAuthenticated={isAuthenticated} />
        );
    }
  };

  return <>{renderPage()}</>;
};

export default App;
