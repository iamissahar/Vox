import { useState, useEffect } from "react";
import { useRouter } from "./hooks/useRouter";
import "./App.css";
import HomePage from "./pages/HomePage";
import LoginPage from "./pages/LoginPage";
import AdminPage from "./pages/AdminPage";
import RoomJoinPage from "./pages/RoomJoinPage";
import RoomPage from "./pages/RoomPage";

export const authState = { user: null };

const S = {
  app: {
    fontFamily: "'DM Sans', 'Helvetica Neue', sans-serif",
    background: "#080808",
    color: "#f0ede8",
    minHeight: "100vh",
    display: "flex",
    flexDirection: "column",
  },
};

function App() {
  const { path, navigate } = useRouter();
  const [user, setUser] = useState(authState.user);

  // Parse room from hash: #/room/VOX-XXXXXX
  const roomMatch = path.match(/^#\/room\/(.+)$/);
  const roomId = roomMatch?.[1];

  return (
    <>
      <div style={S.app}>
        {path === "#/" || path === "" ? (
          <HomePage navigate={navigate} />
        ) : path === "#/login" ? (
          <LoginPage navigate={navigate} onLogin={setUser} />
        ) : path === "#/admin" ? (
          user ? (
            <AdminPage
              navigate={navigate}
              user={user}
              onLogout={() => setUser(null)}
            />
          ) : (
            <LoginPage
              navigate={navigate}
              onLogin={(u) => {
                setUser(u);
                navigate("#/admin");
              }}
            />
          )
        ) : path === "#/room" ? (
          <RoomJoinPage navigate={navigate} />
        ) : roomId ? (
          <RoomPage roomId={roomId} navigate={navigate} />
        ) : (
          <HomePage navigate={navigate} />
        )}
      </div>
    </>
  );
}

export default App;
