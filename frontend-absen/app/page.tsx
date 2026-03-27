"use client";
import React, { useEffect, useRef, useState } from "react";
import * as faceapi from "face-api.js";
import axios from "axios";

export default function AbsensiPage() {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [isModelLoaded, setIsModelLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: "success" | "error" | "info" } | null>(null);

  useEffect(() => {
    const loadModels = async () => {
      setMessage({ text: "Memuat model AI...", type: "info" });
      try {
        await Promise.all([
          faceapi.nets.tinyFaceDetector.loadFromUri("/models"),
          faceapi.nets.faceLandmark68Net.loadFromUri("/models"),
          faceapi.nets.faceRecognitionNet.loadFromUri("/models"),
        ]);
        setIsModelLoaded(true);
        setMessage(null);
        startVideo();
      } catch (e) { setMessage({ text: "Gagal memuat AI.", type: "error" }); }
    };
    loadModels();
  }, []);

  const startVideo = () => {
    navigator.mediaDevices.getUserMedia({ video: true }).then((stream) => {
      if (videoRef.current) videoRef.current.srcObject = stream;
    }).catch(() => setMessage({ text: "Gagal akses kamera.", type: "error" }));
  };

  const handleAbsen = async (jenis: "masuk" | "keluar") => {
    if (!videoRef.current) return;
    setIsLoading(true); setMessage({ text: "Mendeteksi...", type: "info" });

    try {
      const detection = await faceapi.detectSingleFace(videoRef.current, new faceapi.TinyFaceDetectorOptions()).withFaceLandmarks().withFaceDescriptor();
      if (!detection) throw new Error("Wajah tidak terdeteksi");

      const res = await axios.post(`http://localhost:8080/api/absensi/${jenis}`, {
        descriptor: Array.from(detection.descriptor),
      });
      setMessage({ text: `Berhasil! ${res.data.nama} - ${res.data.waktu}`, type: "success" });
    } catch (err: any) {
      setMessage({ text: err.response?.data?.error || err.message, type: "error" });
    } finally {
      setIsLoading(false);
      setTimeout(() => setMessage(null), 5000);
    }
  };

  return (
    <main className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="bg-white p-6 rounded-2xl shadow-xl w-full max-w-2xl text-center">
        <h1 className="text-3xl font-bold mb-6">Sistem Absensi</h1>
        <div className="relative bg-black rounded-lg overflow-hidden aspect-video mb-6">
          <video ref={videoRef} autoPlay muted className="w-full h-full object-cover" />
        </div>
        {message && <div className={`p-4 mb-6 rounded-lg font-bold ${message.type === "success" ? "bg-green-100 text-green-700" : message.type === "error" ? "bg-red-100 text-red-700" : "bg-blue-100 text-blue-700"}`}>{message.text}</div>}
        <div className="flex gap-4 justify-center">
          <button onClick={() => handleAbsen("masuk")} disabled={!isModelLoaded || isLoading} className="px-8 py-3 bg-blue-600 text-white rounded-xl disabled:opacity-50">Absen Masuk</button>
          <button onClick={() => handleAbsen("keluar")} disabled={!isModelLoaded || isLoading} className="px-8 py-3 bg-red-500 text-white rounded-xl disabled:opacity-50">Absen Keluar</button>
        </div>
      </div>
    </main>
  );
}
