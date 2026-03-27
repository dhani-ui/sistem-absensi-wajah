package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pgvector/pgvector-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ==============================================================================
// 1. KONFIGURASI & VARIABEL GLOBAL
// ==============================================================================

var db *gorm.DB
var jwtKey = []byte("rahasia_super_aman_untuk_admin_123") // Ganti secret key jika rilis production

// ==============================================================================
// 2. STRUKTUR DATA (MODEL & REQUEST/RESPONSE)
// ==============================================================================

// Model Database Karyawan
type Karyawan struct {
	ID             uint            `gorm:"primaryKey"`
	Nama           string          `gorm:"size:255;not null"`
	FaceDescriptor pgvector.Vector `gorm:"type:vector(128);not null"`
	CreatedAt      time.Time
}

// Pastikan GORM menggunakan nama tabel tunggal
func (Karyawan) TableName() string {
	return "karyawan"
}

// Model Database Absensi
type Absensi struct {
	ID         uint      `gorm:"primaryKey"`
	KaryawanID uint      `gorm:"index"`
	Nama       string    `gorm:"size:255;not null"`
	Tanggal    time.Time `gorm:"type:date;index"`
	JamMasuk   *string   `gorm:"type:time"` // Pointer (*) agar DB bisa menerima NULL
	JamKeluar  *string   `gorm:"type:time"` // Pointer (*) agar DB bisa menerima NULL
	Keterangan string    `gorm:"size:50"`
}

func (Absensi) TableName() string {
	return "absensi"
}

// Struct untuk Request/Response
type AbsenRequest struct {
	Descriptor []float32 `json:"descriptor" binding:"required"`
}

type MatchResult struct {
	ID       uint
	Nama     string
	Distance float64
}

type LoginCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Nama       string    `json:"nama" binding:"required"`
	Descriptor []float32 `json:"descriptor" binding:"required"`
}

// ==============================================================================
// 3. FUNGSI UTAMA (MAIN)
// ==============================================================================

func main() {
	// Setup Koneksi Database
	// (Ganti user/password/dbname jika berbeda)
	dsn := "host=localhost user=postgres password=rahasia dbname=absenwajah port=5432 sslmode=disable"
	var err error
	
	// Gunakan NamingStrategy agar GORM tidak menjamakkan nama tabel (tidak ditambah huruf 's')
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, 
		},
	})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}

	// AutoMigrate dinonaktifkan karena tabel dibuat manual via raw SQL di Termux
	// db.AutoMigrate(&Karyawan{}, &Absensi{})
	
	log.Println("Database berhasil terkoneksi dan siap digunakan.")

	// Setup Router Gin
	r := gin.Default()

	// CORS Middleware (Wajib agar Next.js bisa menembak API)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// --- Routes Publik (Guest) ---
	r.POST("/api/absensi/masuk", handleAbsenMasuk)
	r.POST("/api/absensi/keluar", handleAbsenKeluar)

	// --- Routes Admin ---
	admin := r.Group("/api/admin")
	admin.POST("/login", handleAdminLogin)
	
	// Gunakan JWT Middleware untuk rute di bawah ini
	admin.Use(JWTMiddleware())
	{
		admin.POST("/register-wajah", handleRegisterWajah)
		admin.GET("/laporan", handleGetLaporanAbsensi)
	}

	log.Println("Server backend berjalan di http://localhost:8080")
	r.Run(":8080")
}

// ==============================================================================
// 4. MIDDLEWARE JWT
// ==============================================================================

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak ditemukan"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode penandatanganan tidak valid")
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak valid atau kedaluwarsa"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ==============================================================================
// 5. HANDLER / CONTROLLERS
// ==============================================================================

// --- Handler Admin Login ---
func handleAdminLogin(c *gin.Context) {
	var creds LoginCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid"})
		return
	}

	if creds.Username != "admin" || creds.Password != "admin123" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	claims := jwt.MapClaims{
		"username": creds.Username,
		"role":     "admin",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)

	c.JSON(http.StatusOK, gin.H{"message": "Login berhasil", "token": tokenString})
}

// --- Handler Registrasi Wajah Baru ---
func handleRegisterWajah(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Descriptor) != 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data wajah tidak valid"})
		return
	}

	karyawanBaru := Karyawan{
		Nama:           req.Nama,
		FaceDescriptor: pgvector.NewVector(req.Descriptor),
		CreatedAt:      time.Now(),
	}

	if err := db.Create(&karyawanBaru).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wajah karyawan berhasil didaftarkan", "nama": req.Nama})
}

// --- Handler Absen Masuk ---
func handleAbsenMasuk(c *gin.Context) {
	var req AbsenRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Descriptor) != 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data wajah tidak valid"})
		return
	}

	inputVector := pgvector.NewVector(req.Descriptor)
	var match MatchResult

	// Cari 1 wajah terdekat dengan toleransi 0.6
	db.Raw(`
		SELECT id, nama, face_descriptor <-> ? AS distance 
		FROM karyawan 
		ORDER BY face_descriptor <-> ? 
		LIMIT 1
	`, inputVector, inputVector).Scan(&match)

	if match.ID == 0 || match.Distance > 0.6 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wajah tidak dikenali atau belum terdaftar"})
		return
	}

	now := time.Now()
	tanggalHariIni := now.Format("2006-01-02")
	jamSekarang := now.Format("15:04:05")

	var absen Absensi
	errCek := db.Where("karyawan_id = ? AND tanggal = ?", match.ID, tanggalHariIni).First(&absen).Error

	if errCek == nil {
		jamM := "-"
		if absen.JamMasuk != nil {
			jamM = *absen.JamMasuk
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah absen masuk hari ini jam " + jamM})
		return
	}

	t, _ := time.Parse("2006-01-02", tanggalHariIni)
	absenBaru := Absensi{
		KaryawanID: match.ID,
		Nama:       match.Nama,
		Tanggal:    t,
		JamMasuk:   &jamSekarang, // Kirim sebagai pointer (alamat memori)
		Keterangan: "Hadir",
	}

	if err := db.Create(&absenBaru).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data absensi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Absen masuk berhasil", "nama": match.Nama, "waktu": jamSekarang})
}

// --- Handler Absen Keluar ---
func handleAbsenKeluar(c *gin.Context) {
	var req AbsenRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Descriptor) != 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data wajah tidak valid"})
		return
	}

	inputVector := pgvector.NewVector(req.Descriptor)
	var match MatchResult

	db.Raw(`
		SELECT id, nama, face_descriptor <-> ? AS distance 
		FROM karyawan 
		ORDER BY face_descriptor <-> ? 
		LIMIT 1
	`, inputVector, inputVector).Scan(&match)

	if match.ID == 0 || match.Distance > 0.6 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wajah tidak dikenali"})
		return
	}

	now := time.Now()
	tanggalHariIni := now.Format("2006-01-02")
	jamSekarang := now.Format("15:04:05")

	var absen Absensi
	errCek := db.Where("karyawan_id = ? AND tanggal = ?", match.ID, tanggalHariIni).First(&absen).Error

	if errCek == gorm.ErrRecordNotFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda belum absen masuk hari ini. Silakan absen masuk dulu."})
		return
	}

	// Cek apakah jam keluar sudah ada nilainya (tidak NULL)
	if absen.JamKeluar != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Anda sudah absen keluar pada jam " + *absen.JamKeluar})
		return
	}

	// Update record
	if err := db.Model(&absen).Update("jam_keluar", jamSekarang).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update absen keluar: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Absen keluar berhasil", "nama": match.Nama, "waktu": jamSekarang})
}

// --- Handler Report (Admin) ---
func handleGetLaporanAbsensi(c *gin.Context) {
	tanggalReq := c.Query("tanggal")
	var daftarAbsensi []Absensi
	query := db.Order("tanggal DESC, jam_masuk DESC").Limit(500)

	if tanggalReq != "" {
		if parsedDate, err := time.Parse("2006-01-02", tanggalReq); err == nil {
			query = query.Where("tanggal = ?", parsedDate)
		}
	}

	if err := query.Find(&daftarAbsensi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca laporan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": daftarAbsensi})
}
