package auth

type RegisterRequest struct {
	Name       string `json:"nama" validate:"required,max=200"`
	Password   string `json:"kata_sandi" validate:"required,min=6,max=72"`
	Phone      string `json:"no_telp" validate:"required,number,min=8,max=15"`
	BirthDate  string `json:"tanggal_Lahir" validate:"required"`
	Occupation string `json:"pekerjaan" validate:"required,max=100"`
	Email      string `json:"email" validate:"required,email,max=255"`
	ProvinceID string `json:"id_provinsi" validate:"required,number,max=10"`
	CityID     string `json:"id_kota" validate:"required,number,max=10"`
}

type LoginRequest struct {
	Phone    string `json:"no_telp" validate:"required,number,max=20"`
	Password string `json:"kata_sandi" validate:"required,max=72"`
}

type LocationRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LoginResponse struct {
	Name       string      `json:"nama"`
	Phone      string      `json:"no_telp"`
	BirthDate  string      `json:"tanggal_Lahir"`
	Occupation string      `json:"pekerjaan"`
	Email      string      `json:"email"`
	Province   LocationRef `json:"id_provinsi"`
	City       LocationRef `json:"id_kota"`
	Token      string      `json:"token"`
}
