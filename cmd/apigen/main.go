// Generates api/schema.graphql and api/openapi.yaml for frontend consumption.
package main

import (
	"log"
	"os"
)

func main() {
	if err := os.MkdirAll("api", 0755); err != nil {
		log.Fatalf("mkdir api: %v", err)
	}

	copyGraphQL()
	writeOpenAPI()

	log.Println("generated: api/schema.graphql, api/openapi.yaml")
}

func copyGraphQL() {
	data, err := os.ReadFile("internal/graph/schema.graphqls")
	if err != nil {
		log.Fatalf("read schema: %v", err)
	}
	if err := os.WriteFile("api/schema.graphql", data, 0644); err != nil {
		log.Fatalf("write schema: %v", err)
	}
}

func writeOpenAPI() {
	spec := `openapi: "3.0.3"
info:
  title: Artline API
  version: "1.0"
  description: |
    REST endpoints for auth and health.
    All other operations go through GraphQL at POST /graphql.

servers:
  - url: http://localhost:8080
    description: Local dev

paths:
  /health:
    get:
      summary: Health check
      operationId: health
      tags: [system]
      responses:
        "200":
          description: OK
          content:
            text/plain:
              schema:
                type: string
                example: ok

  /auth/google/login:
    get:
      summary: Redirect to Google OAuth login
      operationId: authGoogleLogin
      tags: [auth]
      responses:
        "307":
          description: Redirect to Google

  /auth/google/callback:
    get:
      summary: Google OAuth callback — sets session cookie and redirects to frontend
      operationId: authGoogleCallback
      tags: [auth]
      parameters:
        - name: code
          in: query
          required: true
          schema:
            type: string
        - name: state
          in: query
          required: true
          schema:
            type: string
      responses:
        "307":
          description: Redirect to frontend
        "400":
          description: Invalid state
        "500":
          description: Token exchange or DB error

  /auth/logout:
    post:
      summary: Clear session cookie
      operationId: authLogout
      tags: [auth]
      responses:
        "200":
          description: Session cleared

  /auth/me:
    get:
      summary: Returns the currently authenticated user
      operationId: authMe
      tags: [auth]
      security:
        - cookieAuth: []
      responses:
        "200":
          description: Authenticated user
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "401":
          description: Not authenticated

  /graphql:
    post:
      summary: GraphQL endpoint
      operationId: graphql
      tags: [graphql]
      security:
        - cookieAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [query]
              properties:
                query:
                  type: string
                variables:
                  type: object
                operationName:
                  type: string
      responses:
        "200":
          description: GraphQL response (errors are in the response body, not HTTP status)
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                  errors:
                    type: array
                    items:
                      type: object
                      properties:
                        message:
                          type: string

components:
  securitySchemes:
    cookieAuth:
      type: apiKey
      in: cookie
      name: artline_session

  schemas:
    User:
      type: object
      required: [id, email]
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
        displayName:
          type: string
          nullable: true
`
	if err := os.WriteFile("api/openapi.yaml", []byte(spec), 0644); err != nil {
		log.Fatalf("write openapi: %v", err)
	}
}
