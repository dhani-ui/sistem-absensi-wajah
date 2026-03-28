sistem absensi face recognition

golang backend
nextjs frontend
postgresql

-- Buat database
CREATE DATABASE absenwajah;
\c absenwajah;

CREATE EXTENSION vector; 

-- Tabel Karyawan & Data Wajah
CREATE TABLE karyawan (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    -- Jika pakai pgvector: face_descriptor vector(128)
    -- Jika standar: simpan array 128D sebagai JSONB
    face_descriptor JSONB NOT NULL, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel Absensi
CREATE TABLE absensi (
    id SERIAL PRIMARY KEY,
    karyawan_id INT REFERENCES karyawan(id),
    nama VARCHAR(255) NOT NULL,
    tanggal DATE NOT NULL,
    jam_masuk TIME,
    jam_keluar TIME,
    keterangan VARCHAR(50) DEFAULT 'Hadir'
);


CREATE INDEX idx_absensi_tanggal ON absensi(tanggal);
CREATE INDEX idx_absensi_karyawan ON absensi(karyawan_id);

GRANT USAGE, SELECT ON SEQUENCE barang_id_seq TO postgres;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;


backend path :

go run main.go

frontend path :

npm i

npm run dev -- --webpack



fitur :


guest

absen muka

admin;

/admin/login

registrasi

laporan
