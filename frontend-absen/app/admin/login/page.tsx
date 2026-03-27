"use client";
import React, { useState } from "react";
import { useRouter } from "next/navigation";
import axios from "axios";
import Cookies from "js-cookie";

export default function AdminLoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await axios.post("http://localhost:8080/api/admin/login", { username, password });
      Cookies.set("admin_token", res.data.token, { expires: 1 });
      router.push("/admin/report");
    } catch (err) { alert("Login gagal!"); }
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-gray-100">
      <form onSubmit={handleLogin} className="bg-white p-8 rounded-xl shadow-md w-96 flex flex-col gap-4">
        <h1 className="text-2xl font-bold text-center">Login Admin</h1>
        <input type="text" placeholder="Username" onChange={e => setUsername(e.target.value)} className="border p-2 rounded" required />
        <input type="password" placeholder="Password" onChange={e => setPassword(e.target.value)} className="border p-2 rounded" required />
        <button type="submit" className="bg-blue-600 text-white p-2 rounded">Masuk</button>
      </form>
    </main>
  );
}
