"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import axios from "axios";
import Cookies from "js-cookie";

export default function Laporan() {
  const router = useRouter();
  const [data, setData] = useState<any[]>([]);
  const [tgl, setTgl] = useState("");

  useEffect(() => {
    const token = Cookies.get("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }

    const url = tgl ? `http://localhost:8080/api/admin/laporan?tanggal=${tgl}` : "http://localhost:8080/api/admin/laporan";
    axios.get(url, { headers: { Authorization: `Bearer ${token}` } })
      .then(res => setData(res.data.data || []))
      .catch(() => router.push("/admin/login"));
  }, [tgl, router]);

  return (
    <main className="min-h-screen p-10 bg-gray-50">
      <div className="max-w-5xl mx-auto bg-white p-6 rounded-xl shadow">
        <div className="flex justify-between mb-6">
          <h1 className="text-2xl font-bold">Laporan Absensi</h1>
          <div className="flex gap-2">
            <input type="date" value={tgl} onChange={e => setTgl(e.target.value)} className="border p-2 rounded" />
            <button onClick={() => router.push('/admin/register')} className="bg-blue-600 text-white px-4 rounded">Register</button>
            <button onClick={() => { Cookies.remove("admin_token"); router.push('/admin/login') }} className="bg-red-600 text-white px-4 rounded">Logout</button>
          </div>
        </div>
        <table className="w-full text-left border">
          <thead className="bg-gray-100">
            <tr>
              <th className="p-3">Nama</th>
              <th className="p-3">Tgl</th>
              <th className="p-3">Masuk</th>
              <th className="p-3">Keluar</th>
            </tr>
          </thead>
          <tbody>
            {data.map((d, i) => (
              <tr key={i} className="border-b hover:bg-gray-50">
                <td className="p-3 font-bold">{d.Nama}</td>
                <td className="p-3">{new Date(d.Tanggal).toLocaleDateString("id-ID")}</td>
                <td className="p-3 text-green-600 font-bold">{d.JamMasuk || "-"}</td>
                <td className="p-3 text-red-600 font-bold">{d.JamKeluar || "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </main>
  );
}
