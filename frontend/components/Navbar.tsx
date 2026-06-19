"use client";

import { useAuth } from "@/context/auth";
import ThemeToggle from "./ThemeToggle";

export default function Navbar() {
  const { user, logout } = useAuth();
  return (
    <nav className="navbar">
      <span className="brand">✓ Task Manager</span>
      <div className="spacer" />
      <div className="nav-actions">
        <ThemeToggle />
        {user && (
          <>
            <span className="user-email">
              {user.email}
              {user.role === "admin" ? " (admin)" : ""}
            </span>
            <button className="btn" onClick={logout}>
              Log out
            </button>
          </>
        )}
      </div>
    </nav>
  );
}
