package model

type CryptoEncryptMetadata struct {
	GofiberMetadata
	Body struct {
		PlainData       any    `json:"plain_data" validate:"required"`
		ClientPublicKey string `json:"client_public_key" validate:"required"`
		ServerPublicKey string `json:"server_public_key"`
	}
}

type CryptoDecryptMetadata struct {
	GofiberMetadata
	Body struct {
		Encrypted       string `json:"encrypted" validate:"required"`
		ClientPublicKey string `json:"client_public_key" validate:"required"`
		ServerPublicKey string `json:"server_public_key" validate:"required"`
		Nonce           string `json:"nonce" validate:"required"`
	}
}
