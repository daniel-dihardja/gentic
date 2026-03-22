package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SalesData represents restaurant sales metrics
type SalesData struct {
	RestaurantID string    `json:"restaurant_id"`
	Date         string    `json:"date"`
	TotalSales   float64   `json:"total_sales"`
	ItemsSold    []ItemSale `json:"items_sold"`
	CustomerCount int      `json:"customer_count"`
	AverageCheck  float64   `json:"average_check"`
}

// ItemSale represents an individual item's sales
type ItemSale struct {
	ItemID   string  `json:"item_id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

// MenuItem represents a menu item
type MenuItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	PopularityScore float64 `json:"popularity_score"`
}

// RestaurantService handles all backend operations for restaurants
// This service has access to private credentials (DB, API keys, auth tokens)
// Tools never directly access these - they call service methods instead
type RestaurantService struct {
	// Private credentials (would be loaded from env/secrets in real app)
	dbCredentials map[string]string // restaurant_id -> db_password
	apiKeys       map[string]string // service -> api_key

	// Mock data storage (in real app, this would be a database)
	restaurantData map[string]*RestaurantInfo
}

// RestaurantInfo holds restaurant-specific data
type RestaurantInfo struct {
	ID          string
	Name        string
	Cuisine     string
	Owner       string
	Authorized  bool // Whether this restaurant is authorized
}

// NewRestaurantService creates a service with proper credentials
func NewRestaurantService() *RestaurantService {
	service := &RestaurantService{
		dbCredentials: make(map[string]string),
		apiKeys: map[string]string{
			"openai":     "sk_live_openai_key_xyz",
			"dall_e":     "sk_live_dalle_key_xyz",
			"instagram":  "ig_access_token_xyz",
		},
		restaurantData: make(map[string]*RestaurantInfo),
	}

	// Mock restaurant setup
	service.restaurantData["rest_001"] = &RestaurantInfo{
		ID:         "rest_001",
		Name:       "Mario's Pizzeria",
		Cuisine:    "Italian",
		Owner:      "user_42",
		Authorized: true,
	}
	service.restaurantData["rest_002"] = &RestaurantInfo{
		ID:         "rest_002",
		Name:       "Sakura Sushi",
		Cuisine:    "Japanese",
		Owner:      "user_43",
		Authorized: true,
	}

	// Store mock DB credentials (in real app, these would be from secure vault)
	service.dbCredentials["rest_001"] = "db_pass_rest_001"
	service.dbCredentials["rest_002"] = "db_pass_rest_002"

	return service
}

// FetchSalesData retrieves sales data for a restaurant
// This method has access to DB credentials - tools don't
func (s *RestaurantService) FetchSalesData(restaurantID string) (SalesData, error) {
	// Authorization check
	restaurant, exists := s.restaurantData[restaurantID]
	if !exists || !restaurant.Authorized {
		return SalesData{}, fmt.Errorf("unauthorized: restaurant not found or not authorized")
	}

	// In real app: use s.dbCredentials[restaurantID] to query actual database
	// Here we return mock data
	mockData := s.getMockSalesData(restaurantID)
	return mockData, nil
}

// FetchMenuItems retrieves menu items for a restaurant
func (s *RestaurantService) FetchMenuItems(restaurantID string) ([]MenuItem, error) {
	// Authorization check
	restaurant, exists := s.restaurantData[restaurantID]
	if !exists || !restaurant.Authorized {
		return nil, fmt.Errorf("unauthorized: restaurant not found")
	}

	// Return appropriate menu based on cuisine
	return s.getMockMenuItems(restaurant.Cuisine), nil
}

// AnalyzeTrends analyzes sales trends and identifies patterns
// This is business logic that the LLM will use to make decisions
func (s *RestaurantService) AnalyzeTrends(restaurantID string, data SalesData) (map[string]interface{}, error) {
	if _, exists := s.restaurantData[restaurantID]; !exists {
		return nil, fmt.Errorf("restaurant not found")
	}

	// Find top selling items
	var topItems []ItemSale
	for _, item := range data.ItemsSold {
		if item.Quantity > 5 {
			topItems = append(topItems, item)
		}
	}

	// Calculate trends
	avgRevenue := data.TotalSales / float64(data.CustomerCount)

	return map[string]interface{}{
		"top_items":         topItems,
		"total_sales":       data.TotalSales,
		"customer_count":    data.CustomerCount,
		"avg_check":         data.AverageCheck,
		"avg_revenue_per_customer": avgRevenue,
		"trending_category": s.findTrendingCategory(data.ItemsSold),
	}, nil
}

// GeneratePostCopy is a placeholder that would call the LLM
// In real app, this would use OpenAI/Claude
func (s *RestaurantService) GeneratePostCopy(restaurantID string, analysis map[string]interface{}) (string, error) {
	restaurant, exists := s.restaurantData[restaurantID]
	if !exists {
		return "", fmt.Errorf("restaurant not found")
	}

	// Here we would use s.apiKeys["openai"] to call the API
	// For demo, return a template
	topItems := analysis["top_items"].([]ItemSale)
	var itemNames []string
	for _, item := range topItems {
		itemNames = append(itemNames, item.Name)
	}

	copy := fmt.Sprintf(
		"🍽️ Wow! Our guests loved these today at %s! 🔥\n\nTop picks: %s\n\nCome taste what makes us special! 📍\n\n#%s #FoodieLife",
		restaurant.Name,
		strings.Join(itemNames, ", "),
		strings.ReplaceAll(strings.ToLower(restaurant.Cuisine), " ", ""),
	)

	return copy, nil
}

// GenerateImage would call an image generation API
// In real app, this would use DALL-E or similar
func (s *RestaurantService) GenerateImage(restaurantID string, postCopy string) (string, error) {
	if _, exists := s.restaurantData[restaurantID]; !exists {
		return "", fmt.Errorf("restaurant not found")
	}

	// Would use s.apiKeys["dall_e"] here
	// For demo, return a mock image URL
	return "https://images.example.com/generated_post_" + restaurantID + "_" + fmt.Sprint(time.Now().Unix()) + ".jpg", nil
}

// PostToInstagram would post to Instagram
// In real app, this would use Instagram API
func (s *RestaurantService) PostToInstagram(restaurantID string, postCopy string, imageURL string) (map[string]interface{}, error) {
	if _, exists := s.restaurantData[restaurantID]; !exists {
		return nil, fmt.Errorf("restaurant not found")
	}

	// Would use s.apiKeys["instagram"] here
	// For demo, return mock success response
	return map[string]interface{}{
		"post_id":       "ig_post_" + restaurantID + "_123",
		"url":           "https://instagram.com/p/ABC123/",
		"posted_at":     time.Now().Format(time.RFC3339),
		"visibility":    "public",
	}, nil
}

// Helper: Get mock sales data
func (s *RestaurantService) getMockSalesData(restaurantID string) SalesData {
	data := map[string]SalesData{
		"rest_001": {
			RestaurantID: "rest_001",
			Date:         time.Now().Format("2006-01-02"),
			TotalSales:   1250.50,
			CustomerCount: 87,
			AverageCheck:  14.37,
			ItemsSold: []ItemSale{
				{ItemID: "item_001", Name: "Margherita Pizza", Category: "Pizza", Quantity: 24, Revenue: 384.00},
				{ItemID: "item_002", Name: "Carbonara Pasta", Category: "Pasta", Quantity: 18, Revenue: 306.00},
				{ItemID: "item_003", Name: "Tiramisu", Category: "Dessert", Quantity: 12, Revenue: 144.00},
				{ItemID: "item_004", Name: "House Wine", Category: "Beverage", Quantity: 31, Revenue: 248.00},
			},
		},
		"rest_002": {
			RestaurantID: "rest_002",
			Date:         time.Now().Format("2006-01-02"),
			TotalSales:   1890.75,
			CustomerCount: 65,
			AverageCheck:  29.09,
			ItemsSold: []ItemSale{
				{ItemID: "item_201", Name: "Salmon Nigiri (10pc)", Category: "Sushi", Quantity: 15, Revenue: 525.00},
				{ItemID: "item_202", Name: "Dragon Roll", Category: "Sushi", Quantity: 12, Revenue: 408.00},
				{ItemID: "item_203", Name: "Miso Soup", Category: "Soup", Quantity: 22, Revenue: 176.00},
				{ItemID: "item_204", Name: "Sake", Category: "Beverage", Quantity: 28, Revenue: 420.00},
			},
		},
	}

	if data, ok := data[restaurantID]; ok {
		return data
	}
	return SalesData{}
}

// Helper: Get mock menu items
func (s *RestaurantService) getMockMenuItems(cuisine string) []MenuItem {
	if strings.Contains(strings.ToLower(cuisine), "italian") {
		return []MenuItem{
			{ID: "item_001", Name: "Margherita Pizza", Description: "Fresh mozzarella, basil", Category: "Pizza", Price: 16.00, PopularityScore: 0.95},
			{ID: "item_002", Name: "Carbonara Pasta", Description: "Eggs, pecorino, guanciale", Category: "Pasta", Price: 17.00, PopularityScore: 0.92},
			{ID: "item_003", Name: "Tiramisu", Description: "Classic Italian dessert", Category: "Dessert", Price: 12.00, PopularityScore: 0.88},
		}
	}

	// Japanese
	return []MenuItem{
		{ID: "item_201", Name: "Salmon Nigiri (10pc)", Description: "Fresh wild salmon", Category: "Sushi", Price: 35.00, PopularityScore: 0.96},
		{ID: "item_202", Name: "Dragon Roll", Description: "Eel, avocado, cucumber", Category: "Sushi", Price: 34.00, PopularityScore: 0.94},
		{ID: "item_203", Name: "Miso Soup", Description: "Traditional miso broth", Category: "Soup", Price: 8.00, PopularityScore: 0.85},
	}
}

// Helper: Find trending category
func (s *RestaurantService) findTrendingCategory(items []ItemSale) string {
	categoryRevenue := make(map[string]float64)
	for _, item := range items {
		categoryRevenue[item.Category] += item.Revenue
	}

	var topCategory string
	var topRevenue float64
	for category, revenue := range categoryRevenue {
		if revenue > topRevenue {
			topRevenue = revenue
			topCategory = category
		}
	}

	return topCategory
}

// SafeJSONMarshal safely marshals data for tool output
func SafeJSONMarshal(data interface{}) (json.RawMessage, error) {
	bytes, err := json.Marshal(data)
	return json.RawMessage(bytes), err
}
