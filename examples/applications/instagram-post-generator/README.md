# Instagram Post Generator - Production Pattern Example

This example demonstrates the **secure backend service pattern (Pattern A)** for building agentic applications that handle sensitive credentials.

Perfect template for your restaurant Instagram post generation app.

## Architecture

```
┌─────────────────────────────────────────┐
│          Agent Orchestration             │
│   (ReAct loop with tools)                │
└──────────────┬──────────────────────────┘
               │
        (public metadata)
               │
               ▼
┌─────────────────────────────────────────┐
│        Tool Factories (tools.go)         │
│  ├─ fetch_sales_data                    │
│  ├─ analyze_sales_trends                │
│  ├─ fetch_menu_items                    │
│  ├─ generate_post_copy                  │
│  ├─ generate_image                      │
│  └─ post_to_instagram                   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│     RestaurantService (service.go)      │
│  ├─ DB Credentials                      │
│  ├─ API Keys (OpenAI, DALL-E, IG)      │
│  └─ Business Logic                      │
│     ├─ FetchSalesData                   │
│     ├─ AnalyzeTrends                    │
│     ├─ GeneratePostCopy                 │
│     ├─ GenerateImage                    │
│     └─ PostToInstagram                  │
└─────────────────────────────────────────┘
```

## Key Security Properties

✅ **Private Credentials Protected**
- DB passwords, API keys, auth tokens never passed to tools
- Stored only in RestaurantService
- LLM never sees them

✅ **Least Privilege**
- Tools only receive public metadata
- Service enforces authorization
- Each operation validated

✅ **Clear Separation**
- Thin tools (1-2 function calls to service)
- Fat service (business logic + credentials)
- Easy to audit and maintain

## How to Adapt for Your Real App

### 1. Replace Mock Data with Real Data

**In `service.go`, replace mock data:**

```go
// Before: Mock data
func (s *RestaurantService) getMockSalesData(restaurantID string) SalesData {
    return SalesData{...} // hardcoded
}

// After: Real database
func (s *RestaurantService) FetchSalesData(restaurantID string) (SalesData, error) {
    query := "SELECT * FROM sales WHERE restaurant_id = $1"
    rows, err := s.db.Query(query, restaurantID)
    // Parse into SalesData
    return data, err
}
```

### 2. Implement Real LLM Calls

**In `service.go`:**

```go
func (s *RestaurantService) GeneratePostCopy(restaurantID string, analysis map[string]interface{}) (string, error) {
    // Use real LLM API
    resp, err := openai.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: "gpt-4",
            Messages: []openai.ChatCompletionMessage{
                {Role: "system", Content: "You are an Instagram expert..."},
                {Role: "user", Content: fmt.Sprintf("Create post for: %v", analysis)},
            },
            APIKey: s.apiKeys["openai"], // ← Has access to the key!
        },
    )
    return resp.Choices[0].Message.Content, err
}
```

### 3. Connect Real APIs

**Image Generation:**

```go
func (s *RestaurantService) GenerateImage(restaurantID string, postCopy string) (string, error) {
    resp, err := openai.CreateImage(
        context.Background(),
        openai.ImageRequest{
            Prompt: postCopy,
            N:      1,
            Size:   "1024x1024",
            APIKey: s.apiKeys["dall_e"],
        },
    )
    return resp.Data[0].URL, err
}
```

**Instagram Posting:**

```go
func (s *RestaurantService) PostToInstagram(restaurantID string, postCopy string, imageURL string) (map[string]interface{}, error) {
    // Use Instagram Graph API
    resp, err := http.PostForm("https://graph.instagram.com/me/media", url.Values{
        "image_url": {imageURL},
        "caption":   {postCopy},
        "access_token": {s.apiKeys["instagram"]},
    })
    // Parse response
    return result, err
}
```

### 4. Load Real Credentials at Startup

**Replace mock initialization:**

```go
// Before: Hardcoded in code
func NewRestaurantService() *RestaurantService {
    service.apiKeys["openai"] = "sk_live_..."
}

// After: From environment/vault
func NewRestaurantService() (*RestaurantService, error) {
    service := &RestaurantService{
        db: connectToDatabase(),
        apiKeys: map[string]string{
            "openai":    os.Getenv("OPENAI_API_KEY"),
            "dall_e":    os.Getenv("DALLE_API_KEY"),
            "instagram": os.Getenv("INSTAGRAM_ACCESS_TOKEN"),
        },
    }
    return service, nil
}
```

### 5. Add Authorization

**In service methods:**

```go
func (s *RestaurantService) FetchSalesData(restaurantID string) (SalesData, error) {
    // Verify restaurant exists and is authorized
    rest, err := s.db.GetRestaurant(restaurantID)
    if err != nil {
        return SalesData{}, fmt.Errorf("unauthorized: %w", err)
    }

    // Use DB credentials specific to this restaurant
    dbConn := s.createDBConnection(rest.DBHost, rest.DBUser, rest.DBPass)

    // Fetch real data
    return dbConn.GetSalesData(restaurantID, time.Now())
}
```

## Running the Example

```bash
# Build
go build ./examples/instagram-post-generator/

# Run
./examples/instagram-post-generator/main
```

You'll see:
- Mock sales data analysis
- Tool execution flow
- Simulated post generation
- Security confirmation (credentials were never leaked)

## What's Next

1. **Copy this structure** into your real app
2. **Replace mock data** with real database queries
3. **Implement real API calls** to LLM, image generation, Instagram
4. **Load credentials properly** (env vars, secret vault, etc.)
5. **Add error handling** and logging
6. **Add tests** for tools and service methods
7. **Deploy securely** ensuring credentials stay in the service layer

## Key Files

- **`service.go`** - All backend operations + credentials
  - Copy: RestaurantService methods
  - Replace: Mock data with real DB/API calls

- **`tools.go`** - Tool factories
  - Copy: Tool creation pattern (closures over service)
  - Keep: Authorization checks in tools
  - Update: Business logic based on your needs

- **`main.go`** - Agent orchestration
  - Copy: The overall flow
  - Update: Metadata keys for your use case
  - Update: Prompts for your domain

## Security Checklist

- [ ] Private keys stored in service, not metadata
- [ ] Credentials loaded at startup from secure source
- [ ] Tools use SecureMetadata(), not state.Metadata directly
- [ ] Authorization checks in both tools and service
- [ ] No sensitive data in tool outputs
- [ ] Metadata leak detection enabled in production
- [ ] Secrets rotation documented
- [ ] Audit logging for sensitive operations

## For Your Restaurant Post Generator

**Expected metadata:**
```go
Metadata: map[string]interface{}{
    "restaurant_id":    "rest_123",    // Tool accesses
    "user_id":         "user_42",     // For audit
    "request_id":      "req_xyz",     // For tracing
    "_db_password":    "...",         // Protected
    "_openai_key":     "...",         // Protected
    "_instagram_token": "...",        // Protected
}
```

**Tools will:**
1. Fetch today's sales data (using service)
2. Analyze trends automatically
3. Fetch menu items for context
4. Generate engaging post copy (via LLM)
5. Create images (via DALL-E)
6. Post to Instagram (via IG API)

All without ever touching the private credentials!
