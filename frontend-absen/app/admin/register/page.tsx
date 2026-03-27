"use client";
import React, { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import * as faceapi from "face-api.js";
import axios from "axios";
import Cookies from "js-cookie";

export default function RegisterWajah() {
  const router = useRouter();
  const videoRef = useRef<HTMLVideoElement>(null);
  const [nama, setNama] = useState("");

  useEffect(() => {
    if (!Cookies.get("admin_token")) router.push("/admin/login");
    Promise.all([
      faceapi.nets.tinyFaceDetector.loadFromUri("/models"),
      faceapi.nets.faceLandmark68Net.loadFromUri("/models"),
      faceapi.nets.faceRecognitionNet.loadFromUri("/models"),
    ]).then(() => {
      navigator.mediaDevices.getUserMedia({ video: true }).then(s => { if(videoRef.current) videoRef.current.srcObject = s });
    });
  }, [router]);

  const handleRegister = async () => {
    if (!videoRef.current || !nama) return alert("Isi nama dulu!");
    try {
      const detection = await faceapi.detectSingleFace(videoRef.current, new faceapi.TinyFaceDetectorOptions()).withFaceLandmarks().withFaceDescriptor();
      if (!detection) return alert("Wajah tidak terdeteksi");

      await axios.post("http://localhost:8080/api/admin/register-wajah", 
        { nama, descriptor: Array.from(detection.descriptor) },
        { headers: { Authorization: `Bearer ${Cookies.get("admin_token")}` } }
      );
      alert("Berhasil didaftarkan!"); setNama("");
    } catch (err) { alert("Gagal mendaftar"); }
  };

  return (
    <main className="min-h-screen bg-gray-100 flex flex-col items-center py-10">
      <div className="bg-white p-6 rounded-xl shadow-xl w-full max-w-xl text-center">
        <button onClick={() => router.push('/admin/report')} className="mb-4 text-blue-600 underline">Ke Laporan</button>
        <h1 className="text-2xl font-bold mb-4">Register Wajah Baru</h1>
        <input value={nama} onChange={e => setNama(e.target.value)} placeholder="Nama Karyawan" className="w-full p-2 border mb-4 rounded" />
        <video ref={videoRef} autoPlay muted className="w-full bg-black aspect-video rounded mb-4" />
        <button onClick={handleRegister} className="w-full bg-green-600 text-white p-3 rounded font-bold">Daftarkan Wajah</button>
      </div>
    </main>
  );
}
