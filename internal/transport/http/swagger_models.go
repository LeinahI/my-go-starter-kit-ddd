package http

// Swagger models mirror JSON responses for OpenAPI generation only.

type SwaggerHealth struct {
	Status string `json:"status" example:"ok"`
}

type SwaggerHealthEnvelope struct {
	Data SwaggerHealth `json:"data"`
}

type SwaggerProduct struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"Starter Kit"`
	Slug        string `json:"slug" example:"starter-kit"`
	Description string `json:"description,omitempty" example:"Example product"`
	Price       string `json:"price" example:"19.99"`
	Stock       int    `json:"stock" example:"10"`
}

type SwaggerProductListEnvelope struct {
	Data []SwaggerProduct `json:"data"`
}

type SwaggerProductEnvelope struct {
	Data SwaggerProduct `json:"data"`
}

type SwaggerCreateProductRequest struct {
	Name        string `json:"name" example:"Starter Kit"`
	Slug        string `json:"slug" example:"starter-kit"`
	Description string `json:"description" example:"Example product"`
	Price       string `json:"price" example:"19.99"`
	Stock       int    `json:"stock" example:"10"`
}

type SwaggerErrorEnvelope struct {
	Error string `json:"error" example:"not found"`
}
