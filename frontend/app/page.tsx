"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth";

// Entry point: send authenticated users to their tasks, others to login.
export default function Home() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (loading) return;
    router.replace(user ? "/tasks" : "/login");
  }, [user, loading, router]);

  return <div className="spinner" />;
}
