// Package strapi provides a client for interacting with Strapi CMS API.
//
// The Strapi client supports blog post management, content type operations,
// and other CMS-related functionality.
//
// Example usage:
//
//	client := strapi.New(
//		strapi.WithAPIClient(apiClient),
//		strapi.WithBaseURL("https://strapi.example.com"),
//	)
//
//	// Create a blog post
//	post, err := client.CreateBlogPost(ctx, &strapi.CreateBlogPostRequest{
//		Title:   "My First Post",
//		Content: "# Hello World\n\nThis is my first blog post!",
//		Author:  "John Doe",
//		Status:  "draft",
//	})
package strapi
