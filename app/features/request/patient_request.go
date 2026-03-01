package request

type AllergyRequest struct {
	DrugName string `json:"drugName" binding:"required"`
	Reaction string `json:"reaction"`
	Severity string `json:"severity"`
}

type CreatePatientRequest struct {
	IDCard          string           `json:"idCard"`
	FirstName       string           `json:"firstName" binding:"required"`
	LastName        string           `json:"lastName" binding:"required"`
	Phone           string           `json:"phone"`
	Email           string           `json:"email"`
	DateOfBirth     string           `json:"dateOfBirth"`
	Gender          string           `json:"gender"`
	Address         string           `json:"address"`
	Allergies       []AllergyRequest `json:"allergies"`
	ChronicDiseases []string         `json:"chronicDiseases"`
	Notes           string           `json:"notes"`
	PDPAConsent     bool             `json:"pdpaConsent"`
}

type UpdatePatientRequest struct {
	FirstName       string           `json:"firstName"`
	LastName        string           `json:"lastName"`
	Phone           string           `json:"phone"`
	Email           string           `json:"email"`
	DateOfBirth     string           `json:"dateOfBirth"`
	Gender          string           `json:"gender"`
	Address         string           `json:"address"`
	Allergies       []AllergyRequest `json:"allergies"`
	ChronicDiseases []string         `json:"chronicDiseases"`
	Notes           string           `json:"notes"`
}
