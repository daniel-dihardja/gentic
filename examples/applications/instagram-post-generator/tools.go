package main

import (
	"encoding/json"
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
)

// CreatePostGeneratorTools creates all tools for the Instagram post generator
// Tools are created as closures that capture the service instance
// This way: tools receive public metadata only, service holds private credentials
func CreatePostGeneratorTools(service *RestaurantService) []react.Tool {
	return []react.Tool{
		CreateFetchSalesDataTool(service),
		CreateAnalyzeSalesTrendsTool(service),
		CreateFetchMenuItemsTool(service),
		CreateGeneratePostCopyTool(service),
		CreateGenerateImageTool(service),
		CreatePostToInstagramTool(service),
	}
}

// CreateFetchSalesDataTool creates a tool that fetches sales data
// The closure captures the service (which has DB credentials)
func CreateFetchSalesDataTool(service *RestaurantService) react.Tool {
	return react.NewToolWithState(
		"fetch_sales_data",
		"Fetches daily sales data for a restaurant including items sold, total revenue, and customer count",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"restaurant_id": {"type": "string", "description": "The restaurant ID to fetch sales for"}
			},
			"required": ["restaurant_id"]
		}`),
		func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			var params struct {
				RestaurantID string `json:"restaurant_id"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Validate against metadata (ensure they're accessing their own restaurant)
			secure := state.SecureMetadata()
			authorizedRestaurant := secure.GetString("restaurant_id")
			if authorizedRestaurant != "" && authorizedRestaurant != params.RestaurantID {
				return nil, fmt.Errorf("unauthorized: cannot access other restaurant's data")
			}

			// Service has access to DB credentials
			data, err := service.FetchSalesData(params.RestaurantID)
			if err != nil {
				return nil, err
			}

			// Return only necessary data, never expose service internals
			return SafeJSONMarshal(data)
		},
	)
}

// CreateAnalyzeSalesTrendsTool creates a tool that analyzes trends
func CreateAnalyzeSalesTrendsTool(service *RestaurantService) react.Tool {
	return react.NewToolWithState(
		"analyze_sales_trends",
		"Analyzes sales trends to identify top items, popular categories, and customer patterns",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"restaurant_id": {"type": "string", "description": "The restaurant ID"},
				"sales_data": {"type": "object", "description": "The sales data to analyze (from fetch_sales_data tool)"}
			},
			"required": ["restaurant_id", "sales_data"]
		}`),
		func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			var params struct {
				RestaurantID string                 `json:"restaurant_id"`
				SalesData    map[string]interface{} `json:"sales_data"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Convert back to SalesData struct
			bytes, _ := json.Marshal(params.SalesData)
			var salesData SalesData
			if err := json.Unmarshal(bytes, &salesData); err != nil {
				return nil, fmt.Errorf("invalid sales data: %w", err)
			}

			// Service analyzes trends
			analysis, err := service.AnalyzeTrends(params.RestaurantID, salesData)
			if err != nil {
				return nil, err
			}

			return SafeJSONMarshal(analysis)
		},
	)
}

// CreateFetchMenuItemsTool creates a tool that fetches menu items
func CreateFetchMenuItemsTool(service *RestaurantService) react.Tool {
	return react.NewToolWithState(
		"fetch_menu_items",
		"Fetches the restaurant's menu items including descriptions and prices",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"restaurant_id": {"type": "string", "description": "The restaurant ID"}
			},
			"required": ["restaurant_id"]
		}`),
		func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			var params struct {
				RestaurantID string `json:"restaurant_id"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			items, err := service.FetchMenuItems(params.RestaurantID)
			if err != nil {
				return nil, err
			}

			return SafeJSONMarshal(items)
		},
	)
}

// CreateGeneratePostCopyTool creates a tool that generates Instagram post copy
func CreateGeneratePostCopyTool(service *RestaurantService) react.Tool {
	return react.NewToolWithState(
		"generate_post_copy",
		"Generates engaging Instagram post copy based on sales analysis and trends",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"restaurant_id": {"type": "string", "description": "The restaurant ID"},
				"analysis": {"type": "object", "description": "The sales analysis from analyze_sales_trends"}
			},
			"required": ["restaurant_id", "analysis"]
		}`),
		func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			var params struct {
				RestaurantID string                 `json:"restaurant_id"`
				Analysis     map[string]interface{} `json:"analysis"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Service generates copy (would call LLM here with its API key)
			copy, err := service.GeneratePostCopy(params.RestaurantID, params.Analysis)
			if err != nil {
				return nil, err
			}

			return SafeJSONMarshal(map[string]string{
				"post_copy": copy,
			})
		},
	)
}

// CreateGenerateImageTool creates a tool that generates post images
func CreateGenerateImageTool(service *RestaurantService) react.Tool {
	return react.NewToolWithState(
		"generate_image",
		"Generates an Instagram-optimized image for the post using AI image generation",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"restaurant_id": {"type": "string", "description": "The restaurant ID"},
				"post_copy": {"type": "string", "description": "The post copy to base image on"}
			},
			"required": ["restaurant_id", "post_copy"]
		}`),
		func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			var params struct {
				RestaurantID string `json:"restaurant_id"`
				PostCopy     string `json:"post_copy"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Service generates image (would call DALL-E with its API key)
			imageURL, err := service.GenerateImage(params.RestaurantID, params.PostCopy)
			if err != nil {
				return nil, err
			}

			return SafeJSONMarshal(map[string]string{
				"image_url": imageURL,
			})
		},
	)
}

// CreatePostToInstagramTool creates a tool that posts to Instagram
func CreatePostToInstagramTool(service *RestaurantService) react.Tool {
	return react.NewToolWithState(
		"post_to_instagram",
		"Posts the generated content to Instagram with the provided copy and image",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"restaurant_id": {"type": "string", "description": "The restaurant ID"},
				"post_copy": {"type": "string", "description": "The post copy"},
				"image_url": {"type": "string", "description": "The image URL"}
			},
			"required": ["restaurant_id", "post_copy", "image_url"]
		}`),
		func(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
			var params struct {
				RestaurantID string `json:"restaurant_id"`
				PostCopy     string `json:"post_copy"`
				ImageURL     string `json:"image_url"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Service posts to Instagram (would use IG API with its access token)
			result, err := service.PostToInstagram(params.RestaurantID, params.PostCopy, params.ImageURL)
			if err != nil {
				return nil, err
			}

			return SafeJSONMarshal(result)
		},
	)
}
